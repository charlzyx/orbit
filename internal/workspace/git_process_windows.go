//go:build windows

package workspace

import (
	"os/exec"
	"syscall"
)

func hideGitConsole(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
