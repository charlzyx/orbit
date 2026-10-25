package native

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/runtimepath"
)

// ModelCatalogOptions captures the active app configuration before worker I/O.
// A non-nil Environment is complete and does not inherit ambient credentials.
type ModelCatalogOptions struct {
	Binary, Home, PiDir, Workspace string
	Environment                    []string
	EnvOverrides                   map[string]string
}

// RefreshPiModelCatalogWithConfig runs only Pi's catalog refresh command, never
// a shell or installer. It leaves restart/resume and the busy gate to the host.
func RefreshPiModelCatalogWithConfig(ctx context.Context, options ModelCatalogOptions) (CommandOutput, error) {
	if err := preferences.ValidatePiDir(options.PiDir); err != nil {
		return CommandOutput{}, err
	}
	env := options.Environment
	if env == nil {
		env = os.Environ()
	}
	values := map[string]string{}
	for _, variable := range env {
		if key, value, ok := strings.Cut(variable, "="); ok {
			if runtime.GOOS == "windows" {
				key = strings.ToUpper(key)
			}
			values[key] = value
		}
	}
	for key, value := range options.EnvOverrides {
		if runtime.GOOS == "windows" {
			key = strings.ToUpper(key)
		}
		values[key] = value
	}
	if options.Home != "" {
		values["HOME"] = options.Home
		values["USERPROFILE"] = options.Home
	}
	if options.PiDir != "" {
		values["PI_CODING_AGENT_DIR"] = preferences.ResolvePiDir(values["HOME"], options.PiDir)
	} else if options.Home != "" {
		values["PI_CODING_AGENT_DIR"] = filepath.Join(options.Home, ".pi", "agent")
	}
	values["PI_SKIP_VERSION_CHECK"] = "1"
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	env = make([]string, 0, len(keys))
	var secrets []string
	for _, key := range keys {
		env = append(env, key+"="+values[key])
		upper := strings.ToUpper(key)
		for _, marker := range []string{"TOKEN", "SECRET", "PASSWORD", "PASSWD", "CREDENTIAL", "API_KEY", "APIKEY", "ACCESS_KEY", "PRIVATE_KEY", "AUTHORIZATION"} {
			if strings.Contains(upper, marker) && values[key] != "" {
				secrets = append(secrets, values[key])
				break
			}
		}
	}
	binary := options.Binary
	if binary == "" {
		binary = runtimepath.ResolveIn("pi", env, runtime.GOOS)
	} else if !strings.ContainsAny(binary, "/\\") {
		binary = runtimepath.ResolveIn(binary, env, runtime.GOOS)
	}
	if binary == "" {
		return CommandOutput{}, errors.New("configured pi executable is not installed")
	}
	if !filepath.IsAbs(binary) {
		return CommandOutput{}, errors.New("configured pi executable must resolve to an absolute path")
	}
	env = runtimepath.AugmentEnvironment(env, filepath.Dir(binary))
	program, args, err := runtimeProgramIn(binary, []string{"update", "--models", "--no-approve"}, env)
	if err != nil {
		return CommandOutput{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, program, args...)
	cmd.Env = env
	cmd.Dir = options.Workspace
	hideConsole(cmd)
	configureRuntimeCancellation(cmd)
	cmd.WaitDelay = 2 * time.Second
	stdout, stderr := boundedOutput{limit: 1 << 20}, boundedOutput{limit: 1 << 20}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	// One-pass longest-first replacement cannot repeatedly expand a marker when
	// several short configured secrets happen to match its letters.
	sort.Slice(secrets, func(i, j int) bool { return len(secrets[i]) > len(secrets[j]) })
	pairs := make([]string, 0, len(secrets)*2)
	for _, secret := range secrets {
		pairs = append(pairs, secret, "[REDACTED]")
	}
	replacer := strings.NewReplacer(pairs...)
	redact := func(p []byte) []byte {
		text := RedactReportText(replacer.Replace(string(p)), "", nil)
		if len(text) > 1<<20 {
			text = text[:1<<20]
		}
		return []byte(text)
	}
	out := CommandOutput{Stdout: redact(stdout.Bytes()), Stderr: redact(stderr.Bytes()), ExitCode: -1}
	if cmd.ProcessState != nil {
		out.ExitCode = cmd.ProcessState.ExitCode()
	}
	if err != nil {
		if ctx.Err() != nil {
			return out, fmt.Errorf("model catalog refresh: %w", ctx.Err())
		}
		return out, fmt.Errorf("%s: %w", RuntimeFailureDetail(out.Stdout, out.Stderr), err)
	}
	return out, nil
}
