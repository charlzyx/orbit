//go:build darwin || linux

package orbitui

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestAttachmentReadRejectsFIFOReplacementWithoutWriter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "replace.png")
	if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := readAttachmentFileWithOpen(path, 8, func(p string) (*os.File, error) {
		if e := os.Remove(p); e != nil {
			return nil, e
		}
		if e := syscall.Mkfifo(p, 0600); e != nil {
			return nil, e
		}
		return openAttachmentFile(p)
	})
	if !errors.Is(err, errAttachmentNotRegular) {
		t.Fatal("FIFO replacement was not refused", err)
	}
	if _, err := readAttachmentFile(path, 8); !errors.Is(err, errAttachmentNotRegular) {
		t.Fatal("existing FIFO was not refused", err)
	}
}
