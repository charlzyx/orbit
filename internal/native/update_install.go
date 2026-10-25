package native

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const updateHelperFlag = "--orbit-update-install"
const updateReadyEnvironment = "ORBIT_UPDATE_READY_FILE"

type InstallLayout struct{ Root, Executable string }

// DiscoverInstallLayout only recognizes a managed app bundle/prefix, never a
// developer checkout, package-manager installation or arbitrary parent folder.
func DiscoverInstallLayout(executable string) (*InstallLayout, error) {
	if processIsRoot() {
		return nil, errors.New("root-owned installations are check-only")
	}
	real, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return nil, err
	}
	real, err = filepath.Abs(real)
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(real)
	root := filepath.Dir(dir)
	relative := filepath.Join("bin", filepath.Base(real))
	if runtime.GOOS == "darwin" {
		if filepath.Base(dir) != "MacOS" || filepath.Base(root) != "Contents" {
			return nil, errors.New("not running from a managed app bundle")
		}
		root = filepath.Dir(root)
		if !strings.HasSuffix(root, ".app") {
			return nil, errors.New("not running from an app bundle")
		}
		relative = filepath.Join("Contents", "MacOS", filepath.Base(real))
	} else if runtime.GOOS == "windows" {
		root = dir
		relative = filepath.Base(real)
	} else if filepath.Base(dir) != "bin" {
		return nil, errors.New("not running from a managed bin directory")
	}
	probe, err := os.MkdirTemp(filepath.Dir(root), ".orbit-update-probe-")
	if err != nil {
		return nil, err
	}
	_ = os.Remove(probe)
	return &InstallLayout{Root: root, Executable: relative}, nil
}
func DefaultArchiveRoot(version string, layout *InstallLayout) string {
	if runtime.GOOS == "darwin" {
		return filepath.Base(layout.Root)
	}
	arch := "x86_64"
	if runtime.GOARCH == "arm64" {
		arch = "aarch64"
	}
	return "orbit-pi-" + version + "-" + arch + "-unknown-linux-gnu"
}
func validateExecutable(root, relative string) error {
	if relative == "" || filepath.IsAbs(relative) || !filepath.IsLocal(relative) {
		return errors.New("unsafe update executable path")
	}
	info, err := os.Lstat(filepath.Join(root, relative))
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
		return errors.New("staged update is missing its executable")
	}
	return nil
}

// Install hands the verified staged payload to a helper. The caller should
// invoke its normal quit path only after Install returns nil, not before the
// helper has acknowledged validation. It does not itself quit or relaunch UI.
func (u *Updater) Install(ctx context.Context) error {
	if err := u.begin(UpdateInstalling); err != nil {
		return err
	}
	defer u.end()
	ctx = u.operationContext(ctx)
	u.mu.Lock()
	s, layout := u.staged, u.opts.Layout
	u.mu.Unlock()
	if s == nil || layout == nil {
		return u.fail(errors.New("no verified staged update"))
	}
	u.changed()
	if runtime.GOOS == "windows" {
		if err := launch(exec.Command(s.Artifact, "/VERYSILENT", "/SUPPRESSMSGBOXES", "/NORESTART")); err != nil {
			return u.fail(err)
		}
		u.mu.Lock()
		u.staged = nil
		u.mu.Unlock()
		return nil
	}
	if err := validateStaging(layout, s.Directory); err != nil {
		return u.fail(err)
	}
	exe, err := os.Executable()
	if err != nil {
		return u.fail(err)
	}
	ready, err := os.CreateTemp(filepath.Dir(layout.Root), "."+filepath.Base(layout.Root)+".ready-")
	if err != nil {
		return u.fail(err)
	}
	marker := ready.Name()
	ready.Close()
	os.Remove(marker)
	args := []string{updateHelperFlag, "--install-dir", layout.Root, "--staged-dir", s.Directory, "--executable", layout.Executable, "--parent-pid", strconv.Itoa(os.Getpid()), "--ready-file", marker}
	cmd := exec.Command(exe, args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return u.fail(err)
	}
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		return u.fail(err)
	}
	answer := make(chan error, 1)
	go func() {
		reader := bufio.NewReader(io.LimitReader(out, 4096))
		line, e := reader.ReadString('\n')
		if e == nil && strings.TrimSpace(line) != "READY" {
			e = fmt.Errorf("unexpected update helper response: %s", strings.TrimSpace(line))
		}
		answer <- e
	}()
	select {
	case err = <-answer:
	case <-ctx.Done():
		err = ctx.Err()
	}
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return u.fail(err)
	}
	go func() { _ = cmd.Wait() }()
	u.mu.Lock()
	u.staged = nil
	u.mu.Unlock()
	return nil
}
func validateStaging(layout *InstallLayout, staged string) error {
	root, err := filepath.Abs(layout.Root)
	if err != nil {
		return err
	}
	stage, err := filepath.Abs(staged)
	if err != nil {
		return err
	}
	if filepath.Dir(root) != filepath.Dir(stage) || !strings.HasPrefix(filepath.Base(stage), "."+filepath.Base(root)+".update-") {
		return errors.New("staged update is not a namespaced sibling of install")
	}
	for _, p := range []string{root, stage} {
		info, err := os.Lstat(p)
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("install and stage must be real directories")
		}
	}
	return validateExecutable(stage, layout.Executable)
}

// RunInstallHelper must be called before starting MyGo. A malformed invocation
// is handled as failure instead of accidentally opening a second app window.
func RunInstallHelper(args []string) (handled bool, exitCode int) {
	if len(args) == 0 || args[0] != updateHelperFlag {
		return false, 0
	}
	if runtime.GOOS == "windows" {
		return true, 1
	}
	values := map[string]string{}
	if len(args)%2 != 1 {
		return true, 1
	}
	for i := 1; i < len(args); i += 2 {
		if _, ok := values[args[i]]; ok {
			return true, 1
		}
		switch args[i] {
		case "--install-dir", "--staged-dir", "--executable", "--parent-pid", "--ready-file":
			values[args[i]] = args[i+1]
		default:
			return true, 1
		}
	}
	layout := InstallLayout{Root: values["--install-dir"], Executable: values["--executable"]}
	stage, ready := values["--staged-dir"], values["--ready-file"]
	pid, err := strconv.Atoi(values["--parent-pid"])
	if err != nil || pid != os.Getppid() || pid <= 1 || validateStaging(&layout, stage) != nil || filepath.Dir(ready) != filepath.Dir(layout.Root) || !strings.HasPrefix(filepath.Base(ready), "."+filepath.Base(layout.Root)+".ready-") {
		return true, 1
	}
	fmt.Fprintln(os.Stdout, "READY")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err = waitForProcessExit(ctx, pid); err != nil {
		writeUpdateError(layout.Root, err)
		return true, 1
	}
	if err = swapAndRelaunch(layout, stage, ready); err != nil {
		writeUpdateError(layout.Root, err)
		return true, 1
	}
	return true, 0
}
func waitForProcessExit(ctx context.Context, pid int) error {
	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()
	for processAlive(pid) {
		select {
		case <-ctx.Done():
			return fmt.Errorf("old app did not exit: %w", ctx.Err())
		case <-tick.C:
		}
	}
	return nil
}
func writeUpdateError(root string, err error) {
	fmt.Fprintln(os.Stderr, "Orbit updater:", err)
	_ = os.WriteFile(filepath.Join(filepath.Dir(root), "."+filepath.Base(root)+".update-error"), []byte(err.Error()), 0600)
}
func swapAndRelaunch(layout InstallLayout, stage, ready string) error {
	backup, err := os.MkdirTemp(filepath.Dir(layout.Root), "."+filepath.Base(layout.Root)+".rollback-")
	if err != nil {
		return err
	}
	_ = os.Remove(backup)
	spawn := func(marker string) (*exec.Cmd, error) {
		cmd := exec.Command(filepath.Join(layout.Root, layout.Executable))
		cmd.Env = append(os.Environ(), updateReadyEnvironment+"="+marker)
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		return cmd, nil
	}
	if err = os.Rename(layout.Root, backup); err != nil {
		return err
	}
	if err = os.Rename(stage, layout.Root); err != nil {
		_ = os.Rename(backup, layout.Root)
		_, _ = spawn("")
		return err
	}
	syncDirectory(filepath.Dir(layout.Root))
	rollback := func(cause error) error {
		failed, err := os.MkdirTemp(filepath.Dir(layout.Root), "."+filepath.Base(layout.Root)+".failed-")
		if err != nil {
			return fmt.Errorf("%w; cannot reserve recovery directory: %v", cause, err)
		}
		os.Remove(failed)
		if err = os.Rename(layout.Root, failed); err != nil {
			return fmt.Errorf("%w; cannot park failed build: %v", cause, err)
		}
		if err = os.Rename(backup, layout.Root); err != nil {
			_ = os.Rename(failed, layout.Root)
			return fmt.Errorf("%w; rollback failed: %v", cause, err)
		}
		_, restartErr := spawn("")
		_ = os.RemoveAll(failed)
		syncDirectory(filepath.Dir(layout.Root))
		if restartErr != nil {
			return fmt.Errorf("%w; original build relaunch failed: %v", cause, restartErr)
		}
		return cause
	}
	cmd, err := spawn(ready)
	if err != nil {
		return rollback(err)
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	guard := time.NewTimer(30 * time.Second)
	defer guard.Stop()
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		if _, err = os.Stat(ready); err == nil {
			_ = os.Remove(ready)
			_ = os.RemoveAll(backup)
			syncDirectory(filepath.Dir(layout.Root))
			return nil
		}
		select {
		case err := <-exited:
			return rollback(fmt.Errorf("new build exited before opening a window: %v", err))
		case <-guard.C:
			fmt.Fprintln(os.Stderr, "Orbit updater: new build has not acknowledged startup; retained rollback at", backup)
			return nil
		case <-tick.C:
		}
	}
}

// SignalUpdateReady is called only once the real main window is ready. A
// timeout leaves the rollback in place; it never fabricates a successful boot.
func SignalUpdateReady() error {
	marker := os.Getenv(updateReadyEnvironment)
	if marker == "" {
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}
	root := filepath.Dir(filepath.Dir(exe))
	if runtime.GOOS == "darwin" {
		if filepath.Base(filepath.Dir(exe)) != "MacOS" || filepath.Base(root) != "Contents" {
			return errors.New("update ready marker outside managed install")
		}
		root = filepath.Dir(root)
		if !strings.HasSuffix(root, ".app") {
			return errors.New("update ready marker outside app bundle")
		}
	} else if filepath.Base(filepath.Dir(exe)) != "bin" {
		return errors.New("update ready marker outside managed install")
	}
	if filepath.Dir(marker) != filepath.Dir(root) || !strings.HasPrefix(filepath.Base(marker), "."+filepath.Base(root)+".ready-") {
		return errors.New("invalid update ready marker scope")
	}

	f, err := os.OpenFile(marker, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = f.WriteString("ready\n")
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	os.Unsetenv(updateReadyEnvironment)
	if err != nil {
		return err
	}
	return closeErr
}
