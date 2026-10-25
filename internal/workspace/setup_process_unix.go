//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package workspace

import (
	"os"
	"os/exec"
	"syscall"
)

// Cancelled setup/Git/gh commands must also terminate descendants keeping
// output pipes open or continuing work after the UI has cancelled them.
func configureSetupCancellation(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		e := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if e == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return e
	}
}
