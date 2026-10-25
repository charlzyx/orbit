package orbitui

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAttachmentReadActualBytesBound(t *testing.T) {
	for _, n := range []int{0, 7, 8, 9, 32} {
		t.Run(strings.Repeat("x", n), func(t *testing.T) {
			raw, err := readAttachmentBytes(bytes.NewReader(bytes.Repeat([]byte{'a'}, n)), 8)
			if n <= 8 {
				if err != nil || len(raw) != n {
					t.Fatal("within limit rejected", len(raw), err)
				}
			} else if err == nil || raw != nil {
				t.Fatal("over-limit input accepted", len(raw), err)
			}
		})
	}
}

// A stale small file stat cannot bound bytes supplied after it. Simulate growth
// with a reader that keeps producing, and prove no read beyond limit+1 is needed.
type growingAttachmentReader struct{ read int }

func (r *growingAttachmentReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'g'
	}
	r.read += len(p)
	return len(p), nil
}
func TestAttachmentReadGrowthConsumesOnlySentinel(t *testing.T) {
	r := &growingAttachmentReader{}
	// A finite envelope keeps the intentionally broken RED implementation safe.
	raw, err := readAttachmentBytes(&finiteGrowthReader{r: r, left: 64}, 8)
	if err == nil || raw != nil || r.read != 9 {
		t.Fatal("growth was not bounded by actual bytes", len(raw), err, r.read)
	}
}

type finiteGrowthReader struct {
	r    *growingAttachmentReader
	left int
}

func (r *finiteGrowthReader) Read(p []byte) (int, error) {
	if r.left == 0 {
		return 0, io.EOF
	}
	if len(p) > r.left {
		p = p[:r.left]
	}
	n, e := r.r.Read(p)
	r.left -= n
	return n, e
}

func TestAttachmentReadFilePolicyAndGrowth(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "image.png")
	if err := os.WriteFile(path, []byte("12345678"), 0600); err != nil {
		t.Fatal(err)
	}
	raw, err := readAttachmentFile(path, 8)
	if err != nil || string(raw) != "12345678" {
		t.Fatal("exact limit regular file", string(raw), err)
	}
	link := filepath.Join(dir, "link.png")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if raw, err := readAttachmentFile(link, 8); err != nil || len(raw) != 8 {
		t.Fatal("regular symlink rejected", err)
	}
	if _, err := readAttachmentFile(dir, 8); !errors.Is(err, errAttachmentNotRegular) {
		t.Fatal("directory policy", err)
	}
	_, err = readAttachmentFileWithOpen(path, 8, func(p string) (*os.File, error) {
		if e := os.WriteFile(p, []byte("123456789"), 0600); e != nil {
			return nil, e
		}
		return openAttachmentFile(p)
	})
	if !errors.Is(err, errAttachmentReadLimit) {
		t.Fatal("growth between stat and open accepted", err)
	}
}
