//go:build windows

package workbench

import (
	"os/exec"
	"syscall"
)

// Native Windows process-tree lifetime is a separate packaged-platform gate.
func configureResourceProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
