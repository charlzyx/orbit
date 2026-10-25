//go:build windows

package native

import "syscall"

func processIsRoot() bool { return false }
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	const processQueryLimitedInformation = 0x1000
	h, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return err == syscall.ERROR_ACCESS_DENIED
	}
	defer syscall.CloseHandle(h)
	var code uint32
	if err = syscall.GetExitCodeProcess(h, &code); err != nil {
		return true
	}
	return code == 259 // STILL_ACTIVE
}
func syncDirectory(path string) {}
