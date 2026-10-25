// Orbit is a native Go + MyGo workbench backed directly by Pi RPC.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/appbundle"
	"github.com/charlzyx/orbit/internal/composerui"
	"github.com/charlzyx/orbit/internal/diagnosticlog"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/charlzyx/orbit/internal/transcriptui"
	orbitui "github.com/charlzyx/orbit/internal/ui"
	ghostty "github.com/egoist/mygo/orbit-terminal"
)

var version = "0.2.12-go-dev"

type arguments struct {
	options        orbitui.RunOptions
	version        bool
	smokeReport    string
	smokeExitAfter time.Duration
}

func parseArguments(args []string, output io.Writer) (arguments, error) {
	var a arguments
	flags := flag.NewFlagSet("Orbit", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&a.options.Home, "home", "", "isolated Orbit and Pi configuration home (default: normal Pi config; bundled Orbit prefs stay private)")
	flags.StringVar(&a.options.Workspace, "workspace", "", "workspace directory")
	flags.StringVar(&a.options.Session, "session", "", "Pi JSONL session path to resume")
	flags.StringVar(&a.options.PiBinary, "pi", "", "Pi executable path")
	flags.BoolVar(&a.options.Offline, "offline", false, "start Pi without network model catalog refresh")
	flags.BoolVar(&a.options.NoExtensions, "no-extensions", false, "disable Orbit's bundled Pi extensions")
	flags.BoolVar(&a.version, "version", false, "print build version")
	flags.StringVar(&a.smokeReport, "smoke-report", "", "write a fresh native bundle launch smoke report (opt-in)")
	flags.DurationVar(&a.smokeExitAfter, "smoke-exit-after", 30*time.Second, "native bundle launch smoke timeout, at most 5m")
	if err := flags.Parse(args); err != nil {
		return a, err
	}
	if a.options.Workspace == "" && flags.NArg() > 0 {
		a.options.Workspace = flags.Arg(0)
	}
	return a, nil
}

func main() { os.Exit(run(os.Args[1:])) }
func run(args []string) int {
	if handled, code := native.RunInstallHelper(args); handled {
		return code
	}
	executable, _ := os.Executable()
	layout, bundled := appbundle.Discover(executable, runtime.GOOS)
	if bundled {
		args = appbundle.FinderArguments(args)
	}
	a, err := parseArguments(args, os.Stderr)
	// Version does not initialize the native UI, validate runtime resources,
	// create application state, or modify the process environment.
	if err == nil && a.version {
		fmt.Println("Orbit", version)
		return 0
	}
	started := time.Now()
	var diagnostics *diagnosticlog.Logger
	shutdownResult := diagnosticlog.ResultOK
	defer func() {
		diagnostics.Record(diagnosticlog.EventAppShutdown, shutdownResult, time.Since(started))
		_ = diagnostics.Close()
	}()
	fail := func(err error) int {
		shutdownResult = diagnosticlog.ResultError
		diagnostics.Record(diagnosticlog.EventAppStartup, diagnosticlog.ResultError, time.Since(started))
		fmt.Fprintln(os.Stderr, "Orbit:", err)
		if a.options.Smoke != nil {
			_ = a.options.Smoke.Fail("Orbit could not complete native startup; see process stderr")
		} else if bundled && a.smokeReport == "" {
			if dialogErr := appbundle.ShowStartupError(err); dialogErr != nil {
				fmt.Fprintln(os.Stderr, "Orbit startup dialog:", dialogErr)
			}
		}
		return 1
	}
	if err == flag.ErrHelp {
		return 0
	}
	if err != nil {
		return fail(err)
	}
	o := a.options
	o.Version = version
	if bundled {
		if o.Session != "" {
			o.Session, err = filepath.Abs(o.Session)
			if err != nil {
				return fail(err)
			}
		}
		// Resolve explicit CLI runtime overrides before prepending bundled runtimes.
		if o.PiBinary != "" {
			if !strings.ContainsAny(o.PiBinary, "/\\") {
				o.PiBinary, err = exec.LookPath(o.PiBinary)
			}
			if err == nil {
				o.PiBinary, err = filepath.Abs(o.PiBinary)
			}
			if err != nil {
				return fail(fmt.Errorf("Explicit Pi executable is unavailable: %w", err))
			}
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return fail(err)
		}
		ambient := os.Environ()
		if a.smokeReport != "" {
			ambient = appbundle.SmokeEnvironment(ambient)
			o.Offline = true
		}
		prepared, err := layout.Prepare(home, appbundle.Options{Home: o.Home, Workspace: o.Workspace, PiBinary: o.PiBinary}, ambient)
		if err != nil {
			return fail(err)
		}
		o.Home, o.Workspace, o.PiBinary = prepared.Home, prepared.Workspace, prepared.PiBinary
		o.UserHome, o.PiAgentDir = prepared.UserHome, prepared.PiAgentDir
		o.Bundled = true
		diagnostics = openDiagnostics(o.Home)
		o.Diagnostics = diagnostics
		if a.smokeReport != "" {
			o.Smoke, err = appbundle.NewSmoke(a.smokeReport, a.smokeExitAfter, layout, prepared)
			if err != nil {
				return fail(err)
			}
			a.options.Smoke = o.Smoke
		}
		if err := appbundle.ApplyEnvironment(prepared.Environment); err != nil {
			return fail(err)
		}
		if err := os.Chdir(o.Workspace); err != nil {
			return fail(err)
		}
		// This explicit path was set before any terminal construction. A missing or
		// broken library is a visible error, never an implicit runtime download.
		if err := ghostty.Load(); err != nil {
			return fail(fmt.Errorf("The bundled terminal runtime could not load: %w. Download a fresh copy of Orbit.app", err))
		}
		o.Smoke.GhosttyLoaded()
	} else if a.smokeReport != "" {
		return fail(fmt.Errorf("--smoke-report requires the native macOS app bundle"))
	}
	if !bundled {
		home := o.Home
		if home == "" {
			home, _ = os.UserHomeDir()
		}
		diagnostics = openDiagnostics(home)
		o.Diagnostics = diagnostics
	}
	if err := composerui.RequireCompatibility(); err != nil {
		return fail(err)
	}
	if err := transcriptui.RequireCompatibility(); err != nil {
		return fail(err)
	}
	if err := orbitui.Run(o); err != nil {
		shutdownResult = diagnosticlog.ResultError
		diagnostics.Record(diagnosticlog.EventNativeRun, diagnosticlog.ResultError, time.Since(started))
		// Run owns errors before its event loop. Never try to run MyGo twice after
		// a native backend failure or a completed application lifecycle.
		fmt.Fprintln(os.Stderr, "Orbit:", err)
		_ = o.Smoke.Fail("The native application did not complete startup")
		return 1
	}
	return 0
}

func openDiagnostics(home string) *diagnosticlog.Logger {
	if absolute, err := filepath.Abs(home); err == nil && home != "" {
		home = absolute
	}
	logger, err := diagnosticlog.Open(home, version)
	if err != nil {
		// Deliberately omit error/path details; diagnostics must not expose data
		// or prevent the user from starting the application.
		fmt.Fprintln(os.Stderr, "Orbit: local diagnostic logging is unavailable")
		return nil
	}
	logger.Record(diagnosticlog.EventAppStartup, diagnosticlog.ResultStarted, 0)
	return logger
}
