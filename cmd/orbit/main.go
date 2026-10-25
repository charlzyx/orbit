// Orbit is a native Go + MyGo workbench backed directly by Pi RPC.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/charlzyx/orbit/internal/composerui"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/charlzyx/orbit/internal/transcriptui"
	orbitui "github.com/charlzyx/orbit/internal/ui"
)

var version = "0.2.12-go-dev"

func main() {
	if handled, code := native.RunInstallHelper(os.Args[1:]); handled {
		os.Exit(code)
	}
	var o orbitui.RunOptions
	var showVersion bool
	flag.StringVar(&o.Home, "home", "", "isolated application home (default: user home)")
	flag.StringVar(&o.Workspace, "workspace", "", "workspace directory")
	flag.StringVar(&o.Session, "session", "", "Pi JSONL session path to resume")
	flag.StringVar(&o.PiBinary, "pi", "", "Pi executable path")
	flag.BoolVar(&o.Offline, "offline", false, "start Pi without network model catalog refresh")
	flag.BoolVar(&o.NoExtensions, "no-extensions", false, "disable Orbit's bundled Pi extensions")
	flag.BoolVar(&showVersion, "version", false, "print build version")
	flag.Parse()
	if showVersion {
		fmt.Println("Orbit", version)
		return
	}
	if o.Workspace == "" && flag.NArg() > 0 {
		o.Workspace = flag.Arg(0)
	}
	if err := composerui.RequireCompatibility(); err != nil {
		fmt.Fprintln(os.Stderr, "Orbit:", err)
		os.Exit(1)
	}
	if err := transcriptui.RequireCompatibility(); err != nil {
		fmt.Fprintln(os.Stderr, "Orbit:", err)
		os.Exit(1)
	}
	o.Version = version
	if e := orbitui.Run(o); e != nil {
		fmt.Fprintln(os.Stderr, "Orbit:", e)
		os.Exit(1)
	}
}
