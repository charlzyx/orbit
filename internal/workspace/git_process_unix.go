//go:build !windows

package workspace

import "os/exec"

func hideGitConsole(cmd *exec.Cmd) {}
