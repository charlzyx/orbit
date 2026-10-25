//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package workbench

import (
	"os"
	"os/exec"
	"syscall"
)

// Metadata tools can spawn SSH or registry helpers. Cancelling the probe must
// also stop those descendants rather than merely closing their output pipes.
func configureResourceProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
}
