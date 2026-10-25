//go:build !darwin && !linux && !freebsd && !openbsd && !netbsd && !dragonfly

package workspace

import "os/exec"

func configureSetupCancellation(cmd *exec.Cmd) { hideGitConsole(cmd) }
