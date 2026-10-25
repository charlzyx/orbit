package native

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/runtimepath"
)

const PiInstallScriptURL = "https://pi.dev/install.sh"
const PiInstallCommand = "curl -fsSL https://pi.dev/install.sh | sh"
const PiPackageName = "@earendil-works/pi-coding-agent"
const PiLatestVersionURL = "https://pi.dev/api/latest-version"

type Dependency struct {
	Name, Bin, Path, Version, InstallHint, DetailKey, Error string
	Required, Installed                                     bool
}

func CheckDependencies(ctx context.Context) []Dependency {
	specs := []Dependency{{Name: "pi", Bin: "pi", Required: true, DetailKey: "onboarding.pi_detail"}, {Name: "Node.js", Bin: "node", Required: true, DetailKey: "onboarding.node_detail"}, {Name: "git", Bin: "git", DetailKey: "onboarding.git_detail"}}
	for i := range specs {
		d := &specs[i]
		d.InstallHint = DependencyInstallHint(runtime.GOOS, d.Bin)
		d.Path = LocateExecutable(d.Bin)
		d.Installed = d.Path != ""
		if d.Installed {
			version, err := ExecutableVersion(ctx, d.Path)
			d.Version = version
			if err != nil {
				d.Error = err.Error()
			}
		}
	}
	return specs
}
func AllRequiredInstalled(deps []Dependency) bool { return MissingRequiredCount(deps) == 0 }
func MissingRequiredCount(deps []Dependency) int {
	n := 0
	for _, d := range deps {
		if d.Required && !d.Installed {
			n++
		}
	}
	return n
}
func DependencyInstallHint(goos, bin string) string {
	if bin == "pi" {
		if goos == "darwin" || goos == "linux" {
			return PiInstallCommand
		}
		return "npm install -g --ignore-scripts @earendil-works/pi-coding-agent"
	}
	switch goos {
	case "darwin":
		if bin == "node" {
			return "brew install node"
		}
		return "brew install git"
	case "windows":
		if bin == "node" {
			return "winget install OpenJS.NodeJS.LTS"
		}
		return "winget install Git.Git"
	default:
		if bin == "node" {
			return "sudo apt install nodejs npm"
		}
		return "sudo apt install git"
	}
}
func runtimeDirectories(home string) []string {
	return runtimepath.Directories(os.Getenv("PATH"), home, runtime.GOOS)
}
func executableIn(dir, name, goos string) string   { return runtimepath.ExistingIn(dir, name, goos) }
func LocateExecutable(name string) string          { return runtimepath.Resolve(name) }
func AugmentedRuntimePath(preferred string) string { return runtimepath.AugmentedPath(preferred) }
func runtimeEnvironment(preferred string) []string {
	var base []string
	for _, v := range os.Environ() {
		key, _, _ := strings.Cut(v, "=")
		if !strings.EqualFold(key, "PI_SKIP_VERSION_CHECK") {
			base = append(base, v)
		}
	}
	return append(runtimepath.AugmentEnvironment(base, preferred), "PI_SKIP_VERSION_CHECK=1")
}

// ExecutableVersion is bounded and checks exit status. A missing Node runtime
// is an error, not the misleading "env: node ..." version line.
func ExecutableVersion(ctx context.Context, binary string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	out, err := runRuntimeCommand(ctx, binary, []string{"--version"}, false)
	if err != nil {
		return "", err
	}
	data := out.Stdout
	if len(data) == 0 {
		data = out.Stderr
	}
	line, _, _ := strings.Cut(string(data), "\n")
	line = strings.TrimSpace(line)
	if line == "" {
		return "", errors.New("executable returned no version")
	}
	return line, nil
}

type PiInstallKind string

const (
	PiManaged PiInstallKind = "managed"
	PiNPM     PiInstallKind = "npm"
	PiOther   PiInstallKind = "other"
)

type PiInstallation struct {
	Bin, Version, Root string
	Kind               PiInstallKind
	OnPath             bool
	Error              string
}

func (p PiInstallation) Installed() bool { return p.Bin != "" }
func PiAgentDirectory() string {
	if path := os.Getenv("PI_CODING_AGENT_DIR"); path != "" {
		return path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pi", "agent")
}
func ManagedPiMarkerValid(root string) bool {
	raw, err := os.ReadFile(filepath.Join(root, "managed-install.json"))
	if err != nil || len(raw) > 1<<20 {
		return false
	}
	var marker struct {
		Kind   string `json:"kind"`
		Schema int    `json:"schemaVersion"`
		Layout string `json:"layout"`
	}
	return json.Unmarshal(raw, &marker) == nil && marker.Kind == "pi-managed-install" && marker.Schema == 1 && marker.Layout == "releases-v1"
}
func canonical(path string) string {
	if real, err := filepath.EvalSymlinks(path); err == nil {
		path = real
	}
	absolute, err := filepath.Abs(path)
	if err == nil {
		return absolute
	}
	return path
}
func within(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	return err == nil && (rel == "." || filepath.IsLocal(rel))
}
func ClassifyPiInstallation(bin, agentDir string) (PiInstallKind, string) {
	real := canonical(bin)
	roots := []string{filepath.Join(agentDir, "install")}
	if override := os.Getenv("PI_MANAGED_INSTALL_ROOT"); override != "" {
		roots = append([]string{override}, roots...)
	}
	for _, root := range roots {
		if ManagedPiMarkerValid(root) {
			root = canonical(root)
			if within(real, root) || within(real, filepath.Join(filepath.Dir(root), "bin")) {
				return PiManaged, root
			}
		}
	}
	dir := filepath.Dir(real)
	for range 12 {
		if isPiPackage(dir) {
			return PiNPM, dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if runtime.GOOS == "windows" {
		pkg := filepath.Join(filepath.Dir(real), "node_modules", "@earendil-works", "pi-coding-agent")
		if isPiPackage(pkg) {
			return PiNPM, pkg
		}
	}
	return PiOther, ""
}
func isPiPackage(path string) bool {
	raw, err := os.ReadFile(filepath.Join(path, "package.json"))
	if err != nil || len(raw) > 1<<20 {
		return false
	}
	var p struct {
		Name string `json:"name"`
	}
	return json.Unmarshal(raw, &p) == nil && p.Name == PiPackageName
}
func DetectPiInstallation(ctx context.Context) PiInstallation {
	bin := LocateExecutable("pi")
	p := PiInstallation{Bin: bin, Kind: PiOther}
	if bin == "" {
		return p
	}
	p.Kind, p.Root = ClassifyPiInstallation(bin, PiAgentDirectory())
	real := canonical(bin)
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		candidate := executableIn(dir, "pi", runtime.GOOS)
		if candidate != "" && canonical(candidate) == real {
			p.OnPath = true
			break
		}
	}
	var err error
	p.Version, err = ExecutableVersion(ctx, bin)
	if err != nil {
		p.Error = err.Error()
	}
	return p
}

type PiUpdatePlan string

const (
	PiUpToDate           PiUpdatePlan = "up_to_date"
	PiUpdateManaged      PiUpdatePlan = "update"
	PiMigrateNPM         PiUpdatePlan = "migrate"
	PiUnsupportedManager PiUpdatePlan = "unsupported_package_manager"
	PiUnknownVersion     PiUpdatePlan = "unknown_version"
)

func PlanPiUpdate(installed PiInstallation, latest string) PiUpdatePlan {
	if installed.Version == "" {
		if installed.Installed() {
			return PiUnknownVersion
		}
		return PiUpToDate
	}
	if latest == "" || CompareVersions(latest, installed.Version) <= 0 {
		return PiUpToDate
	}
	switch installed.Kind {
	case PiManaged:
		return PiUpdateManaged
	case PiNPM:
		return PiMigrateNPM
	default:
		return PiUnsupportedManager
	}
}
func LatestPiVersion(ctx context.Context, client *http.Client) (string, error) {
	if os.Getenv("PI_OFFLINE") != "" {
		return "", nil
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, PiLatestVersionURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Orbit-MyGo")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("pi version endpoint returned HTTP %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20+1))
	if err != nil {
		return "", err
	}
	if len(raw) > 1<<20 {
		return "", errors.New("pi version response too large")
	}
	var v struct {
		Version string `json:"version"`
	}
	if err = json.Unmarshal(raw, &v); err != nil {
		return "", err
	}
	version := strings.TrimSpace(v.Version)
	if version == "" {
		return "", errors.New("pi version endpoint returned no version")
	}
	return version, nil
}

type CommandOutput struct {
	Stdout, Stderr []byte
	ExitCode       int
}
type boundedOutput struct {
	bytes.Buffer
	limit int
}

// Mask bytes.Buffer.ReadFrom: os/exec copies pipes through io.Copy, whose
// ReaderFrom optimization would otherwise bypass the bounded Write method.
func (b *boundedOutput) ReadFrom(r io.Reader) (int64, error) {
	return io.Copy(struct{ io.Writer }{b}, r)
}
func (b *boundedOutput) Write(p []byte) (int, error) {
	n := len(p)
	remaining := b.limit - b.Len()
	if remaining > 0 {
		b.Buffer.Write(p[:min(remaining, n)])
	}
	return n, nil
}
func runRuntimeCommand(ctx context.Context, binary string, args []string, detach bool) (CommandOutput, error) {
	program, argv, err := runtimeProgram(binary, args)
	if err != nil {
		return CommandOutput{}, err
	}
	cmd := exec.CommandContext(ctx, program, argv...)
	cmd.Env = runtimeEnvironment(filepath.Dir(binary))
	hideConsole(cmd)
	if detach {
		detachRuntimeProcess(cmd)
		cmd.Env = append(cmd.Env, "TERM=dumb")
	}
	configureRuntimeCancellation(cmd)
	cmd.WaitDelay = 2 * time.Second
	var stdout, stderr boundedOutput
	stdout.limit = 1 << 20
	stderr.limit = 1 << 20
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = nil
	err = cmd.Run()
	result := CommandOutput{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: -1}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	if err != nil {
		return result, fmt.Errorf("%s: %w", RuntimeFailureDetail(result.Stdout, result.Stderr), err)
	}
	return result, nil
}
func RunPiUpdate(ctx context.Context) (CommandOutput, error) {
	p := DetectPiInstallation(ctx)
	if !p.Installed() {
		return CommandOutput{}, errors.New("pi is not installed")
	}
	if p.Kind != PiManaged {
		return CommandOutput{}, errors.New("only an official managed pi install may be updated; npm requires confirmed migration and other package managers are not modified")
	}
	return runRuntimeCommand(ctx, p.Bin, []string{"update", "--no-approve"}, false)
}
func RefreshPiModelCatalog(ctx context.Context) (CommandOutput, error) {
	bin := LocateExecutable("pi")
	if bin == "" {
		return CommandOutput{}, errors.New("pi is not installed")
	}
	return runRuntimeCommand(ctx, bin, []string{"update", "--models", "--no-approve"}, false)
}
func RuntimeFailureDetail(stdout, stderr []byte) string {
	text := stripRuntimeANSI(string(stdout) + string(stderr))
	var last string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		last = line
		if strings.HasPrefix(line, "Error:") {
			if detail := strings.TrimSpace(strings.TrimPrefix(line, "Error:")); detail != "" {
				return detail
			}
		}
	}
	if last != "" {
		return last
	}
	return "pi update failed"
}
func stripRuntimeANSI(s string) string {
	var b strings.Builder
	escape := false
	for _, r := range s {
		if escape {
			if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
				escape = false
			}
			continue
		}
		if r == 27 {
			escape = true
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
