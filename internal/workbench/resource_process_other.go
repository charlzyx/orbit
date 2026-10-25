//go:build !darwin && !linux && !freebsd && !openbsd && !netbsd && !dragonfly && !windows

package workbench

import "os/exec"

func configureResourceProcess(cmd *exec.Cmd) {}
