package orbitui

import (
	"errors"
	"io"
	"os"
)

const attachmentFileByteLimit int64 = 64 << 20

var errAttachmentReadLimit = errors.New("image exceeds attachment size limit")
var errAttachmentNotRegular = errors.New("image attachment must be a regular file")

// The bound is checked on actual bytes, not just a possibly stale file stat.
func readAttachmentBytes(r io.Reader, limit int64) ([]byte, error) {
	if limit < 0 || limit == int64(^uint64(0)>>1) {
		return nil, errAttachmentReadLimit
	}
	raw, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > limit {
		return nil, errAttachmentReadLimit
	}
	return raw, nil
}
func readAttachmentFile(path string, limit int64) ([]byte, error) {
	return readAttachmentFileWithOpen(path, limit, openAttachmentFile)
}
func readAttachmentFileWithOpen(path string, limit int64, open func(string) (*os.File, error)) ([]byte, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, errAttachmentNotRegular
	}
	if st.Size() > limit {
		return nil, errAttachmentReadLimit
	}
	f, err := open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err = f.Stat()
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, errAttachmentNotRegular
	}
	if st.Size() > limit {
		return nil, errAttachmentReadLimit
	}
	return readAttachmentBytes(f, limit)
}
