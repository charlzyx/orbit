//go:build !windows

package native

import (
	"errors"
	"os/exec"
)

func hideConsole(cmd *exec.Cmd) {}
func launchExternalTerminal(program string, args, environment []string) error {
	return errors.New("Windows console launch is unavailable on this platform")
}
