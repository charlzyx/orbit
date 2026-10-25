//go:build !windows

package native

import (
	"os"
	"os/exec"
	"syscall"
)

func runtimeProgram(binary string, args []string) (string, []string, error) { return binary, args, nil }
func detachRuntimeProcess(cmd *exec.Cmd)                                    { cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} }

// Version probes and update commands may be wrappers which spawn children.
// Give only our own child a fresh process group/session so cancellation does
// not leave its descendants holding captured pipes and hanging the UI worker.
func configureRuntimeCancellation(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err == nil {
			return nil
		}
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return cmd.Process.Kill()
	}
}
