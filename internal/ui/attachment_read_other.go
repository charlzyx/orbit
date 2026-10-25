//go:build !darwin && !linux

package orbitui

import "os"

// Both regular-file checks and the actual-byte bound still apply. Unlike the
// Darwin/Linux path, this fallback does not close a special-file replacement
// race at open itself and makes no bounded syscall-completion guarantee.
func openAttachmentFile(path string) (*os.File, error) { return os.Open(path) }
