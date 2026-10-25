//go:build darwin || linux

package orbitui

import (
	"os"
	"syscall"
)

// O_NONBLOCK prevents a regular-to-FIFO replacement between stat and open from
// waiting for a writer. It does not change regular-file reads or promise that
// arbitrary filesystem operations (e.g. network filesystems) are cancellable.
func openAttachmentFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
