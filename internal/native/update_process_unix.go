//go:build !windows

package native

import (
	"errors"
	"os"
	"syscall"
)

func processIsRoot() bool { return os.Geteuid() == 0 }
func processAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
func syncDirectory(path string) {
	if f, err := os.Open(path); err == nil {
		_ = f.Sync()
		_ = f.Close()
	}
}
