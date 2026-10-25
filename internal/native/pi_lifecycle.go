package native

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type PiInstallOutcome string

const (
	PiInstallNew      PiInstallOutcome = "install"
	PiAlreadyManaged  PiInstallOutcome = "already_managed"
	PiInstallConflict PiInstallOutcome = "conflict"
)

func PlanPiInstall(p PiInstallation) PiInstallOutcome {
	if !p.Installed() {
		return PiInstallNew
	}
	if p.Kind == PiManaged {
		return PiAlreadyManaged
	}
	return PiInstallConflict
}
func OfficialPiInstallSupported() bool { return runtime.GOOS == "darwin" || runtime.GOOS == "linux" }

type Invocation struct {
	Program string
	Args    []string
}

func RunInstallerInvocation(ctx context.Context, invocation Invocation) (CommandOutput, error) {
	if invocation.Program == "" {
		return CommandOutput{}, errors.New("installer program is empty")
	}
	return runRuntimeCommand(ctx, invocation.Program, invocation.Args, true)
}

// InstallPi is an explicit UI action. NPM migration must have been confirmed:
// pi's installer removes the old global package after making the managed one.
// Config, auth, sessions and extensions are never deleted or patched by Orbit.
// The downloaded script is a file argument, never an interpolated shell command.
func InstallPi(ctx context.Context, allowNPMMigration bool) (CommandOutput, error) {
	return installPi(ctx, allowNPMMigration, nil)
}
func installPi(ctx context.Context, allowNPMMigration bool, client *http.Client) (CommandOutput, error) {
	if !OfficialPiInstallSupported() {
		return CommandOutput{}, errors.New("the official pi installer supports macOS and Linux only; use the displayed platform install command")
	}
	installed := DetectPiInstallation(ctx)
	switch PlanPiInstall(installed) {
	case PiAlreadyManaged:
		return CommandOutput{}, errors.New("pi is already officially managed; use update instead of making a second installation")
	case PiInstallConflict:
		if installed.Kind != PiNPM {
			return CommandOutput{}, errors.New("a different package manager owns pi; Orbit will not replace it")
		}
		if !allowNPMMigration {
			return CommandOutput{}, errors.New("migration needs confirmation: the official installer removes the existing npm package")
		}
	}
	if os.Getenv("ORBIT_PI_INSTALL_CMD") != "" {
		return CommandOutput{}, errors.New("raw shell installer overrides are not executed; use an explicit program and argument invocation")
	}
	scriptURL := PiInstallScriptURL
	if override := os.Getenv("ORBIT_PI_INSTALL_URL"); override != "" {
		scriptURL = override
	}
	if err := validDownloadURL(scriptURL); err != nil {
		return CommandOutput{}, err
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	clone := *client
	prior := clone.CheckRedirect
	clone.CheckRedirect = func(r *http.Request, via []*http.Request) error {
		if err := validDownloadURL(r.URL.String()); err != nil {
			return err
		}
		if prior != nil {
			return prior(r, via)
		}
		if len(via) > 10 {
			return errors.New("too many installer redirects")
		}
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, scriptURL, nil)
	if err != nil {
		return CommandOutput{}, err
	}
	resp, err := clone.Do(req)
	if err != nil {
		return CommandOutput{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return CommandOutput{}, fmt.Errorf("official installer returned HTTP %d", resp.StatusCode)
	}
	script, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20+1))
	if err != nil {
		return CommandOutput{}, err
	}
	if len(script) == 0 || len(script) > 4<<20 {
		return CommandOutput{}, errors.New("official installer is empty or exceeds the size limit")
	}
	parent, err := EnsureTempDir()
	if err != nil {
		return CommandOutput{}, err
	}
	dir, err := os.MkdirTemp(parent, fmt.Sprintf("update-%d-installer-", os.Getpid()))
	if err != nil {
		return CommandOutput{}, err
	}
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "install.sh")
	if err = os.WriteFile(file, script, 0600); err != nil {
		return CommandOutput{}, err
	}
	return RunInstallerInvocation(ctx, Invocation{Program: "/bin/sh", Args: []string{file}})
}

type PiUpdateEvent struct {
	State, Latest, Before, Detail string
	Installation                  PiInstallation
	Dependencies                  []Dependency
}
type PiLaunchUpdateOptions struct {
	// Busy must include active, parked and pending sessions' running, streaming
	// and compacting state. A nil Busy cannot safely authorize an automatic swap.
	Busy func() bool
	// BeginUpdate atomically verifies idle state and excludes new session/process
	// mutations until release. Busy is only advisory; a missing/rejected guard
	// never authorizes an automatic update. A successful guard returns a nonnil
	// release callback, called exactly once, including cancellation/failure.
	BeginUpdate func() (release func(), ok bool)
	OnEvent     func(PiUpdateEvent)
	Client      *http.Client
}

// CheckPiUpdateOnLaunch reproduces the original once-per-launch orchestration:
// defer initial traffic 3s, skip offline/custom PI_BIN, require an installed
// managed runtime, wait up to 120s for idle, and re-probe actual results.
// NPM migration is offered only; it is never silently performed.
func CheckPiUpdateOnLaunch(ctx context.Context, opts PiLaunchUpdateOptions) {
	if os.Getenv("PI_BIN") != "" || os.Getenv("PI_OFFLINE") != "" {
		return
	}
	go func() {
		if !waitRuntime(ctx, 3*time.Second) {
			return
		}
		installed := DetectPiInstallation(ctx)
		if !installed.Installed() {
			return
		}
		before := installed.Version
		latest, err := LatestPiVersion(ctx, opts.Client)
		if err != nil || latest == "" {
			return
		}
		plan := PlanPiUpdate(installed, latest)
		emit := func(state, detail string) {
			if opts.OnEvent != nil {
				opts.OnEvent(PiUpdateEvent{State: state, Latest: latest, Before: before, Detail: detail, Installation: installed})
			}
		}
		switch plan {
		case PiUpToDate:
			return
		case PiMigrateNPM:
			emit("migration_required", "The official installer removes the old npm package; confirm migration in setup.")
			return
		case PiUnsupportedManager, PiUnknownVersion:
			emit(string(plan), "Orbit will not modify this pi installation.")
			return
		}
		emit("available", "")
		if opts.Busy == nil {
			emit("deferred", "Waiting for a verified idle session state.")
			return
		}
		deadline := time.NewTimer(120 * time.Second)
		defer deadline.Stop()
		tick := time.NewTicker(5 * time.Second)
		defer tick.Stop()
		for opts.Busy() {
			select {
			case <-ctx.Done():
				return
			case <-deadline.C:
				emit("deferred", "The agent is still busy; the update is deferred to the next launch.")
				return
			case <-tick.C:
			}
		}
		// Re-read the installed launcher just before acting, not the stale startup
		// snapshot, and re-check custom/offline overrides changed in-process.
		if os.Getenv("PI_BIN") != "" || os.Getenv("PI_OFFLINE") != "" || opts.Busy() {
			return
		}

		var event PiUpdateEvent
		acquired := runWithPiUpdateGuard(ctx, opts.BeginUpdate, func() {
			emit("updating", "")
			output, updateErr := RunPiUpdate(ctx)
			verification := DetectPiInstallation(ctx)
			deps := CheckDependencies(ctx)
			event = PiUpdateEvent{State: "unchanged", Latest: latest, Before: before, Installation: verification, Dependencies: deps}
			if updateErr != nil {
				event.State = "failed"
				event.Detail = RuntimeFailureDetail(output.Stdout, output.Stderr)
				if strings.TrimSpace(string(output.Stdout)+string(output.Stderr)) == "" {
					event.Detail = updateErr.Error()
				}
			} else if CompareVersions(verification.Version, before) > 0 {
				event.State = "updated"
				if !verification.OnPath {
					event.Detail = "The updated pi is not on the process PATH."
				}
			}
		})
		if !acquired {
			if ctx.Err() == nil {
				emit("deferred", "The runtime maintenance guard was not acquired; no update was started.")
			}
			return
		}

		if opts.OnEvent != nil {
			opts.OnEvent(event)
		}
	}()
}
func waitRuntime(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// A small transaction seam keeps the exclusion contract testable without
// running any installer. Final notifications happen after this releases.
func runWithPiUpdateGuard(ctx context.Context, begin func() (func(), bool), work func()) bool {
	if begin == nil || ctx.Err() != nil {
		return false
	}
	release, ok := begin()
	if release != nil {
		defer release()
	}
	if !ok || release == nil || ctx.Err() != nil {
		return false
	}
	work()
	return true
}
