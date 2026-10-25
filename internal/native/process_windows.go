//go:build windows

package native

import (
	"os"
	"os/exec"
	"syscall"
)

func hideConsole(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
func launchExternalTerminal(program string, args, environment []string) error {
	cmd := exec.Command(program, args...)
	cmd.Env = append(os.Environ(), environment...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000010}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
