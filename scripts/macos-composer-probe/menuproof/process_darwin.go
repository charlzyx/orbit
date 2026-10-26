//go:build darwin && arm64

package menuproof

import (
	"bytes"
	"errors"
	"github.com/ebitengine/purego"
	"golang.org/x/sys/unix"
	"path/filepath"
	"runtime"
	"unsafe"
)

type Process struct {
	PID        int    `json:"pid"`
	Start      uint64 `json:"start_us"`
	Executable string `json:"executable"`
}

// InspectProcess reads only an explicitly supplied process's start identity and
// executable, never its environment, arguments, clipboard, or files.
func InspectProcess(pid int) (Process, error) {
	if pid < 2 {
		return Process{}, ErrBlocked
	}
	before, e := unix.SysctlKinfoProc("kern.proc.pid", pid)
	if e != nil {
		return Process{}, e
	}
	if int(before.Proc.P_pid) != pid || before.Proc.P_starttime.Sec <= 0 {
		return Process{}, ErrBlocked
	}
	lib, e := purego.Dlopen("/usr/lib/libproc.dylib", purego.RTLD_NOW|purego.RTLD_LOCAL)
	if e != nil {
		return Process{}, e
	}
	defer purego.Dlclose(lib)
	var path func(int32, unsafe.Pointer, uint32) int32
	sym, e := purego.Dlsym(lib, "proc_pidpath")
	if e != nil {
		return Process{}, e
	}
	purego.RegisterFunc(&path, sym)
	b := make([]byte, 4096)
	n := path(int32(pid), unsafe.Pointer(&b[0]), uint32(len(b)))
	runtime.KeepAlive(b)
	if n <= 0 || int(n) >= len(b) {
		return Process{}, errors.New("process path unavailable/bounded")
	}
	exe := string(bytes.TrimRight(b[:n], "\x00"))
	if !filepath.IsAbs(exe) {
		return Process{}, ErrBlocked
	}
	after, e := unix.SysctlKinfoProc("kern.proc.pid", pid)
	if e != nil || after.Proc.P_pid != before.Proc.P_pid || after.Proc.P_starttime != before.Proc.P_starttime {
		return Process{}, ErrBlocked
	}
	return Process{pid, uint64(before.Proc.P_starttime.Sec)*1_000_000 + uint64(before.Proc.P_starttime.Usec), exe}, nil
}
