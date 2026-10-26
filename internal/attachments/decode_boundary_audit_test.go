package attachments

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"testing"
)

// Header only; deliberately never call image.Decode or MyGo DecodeBitmap with
// these dimensions. This proves the existing Prepare passthrough decision only.
func auditPNGHeader(w, h uint32) []byte {
	b := append([]byte(nil), []byte("\x89PNG\r\n\x1a\n")...)
	chunk := make([]byte, 17)
	copy(chunk, "IHDR")
	binary.BigEndian.PutUint32(chunk[4:8], w)
	binary.BigEndian.PutUint32(chunk[8:12], h)
	chunk[12] = 8
	chunk[13] = 6
	b = binary.BigEndian.AppendUint32(b, 13)
	b = append(b, chunk...)
	b = binary.BigEndian.AppendUint32(b, crc32.ChecksumIEEE(chunk))
	return b
}
func TestAuditOversizedPNGPreparePreservesRawHeader(t *testing.T) {
	raw := auditPNGHeader(32768, 32768)
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || cfg.Width != 32768 || cfg.Height != 32768 {
		t.Fatal("safe tinyheader not recognized", cfg, err)
	}
	got := Prepare(raw, "image/png")
	if !bytes.Equal(got.Bytes, raw) || got.Compressed || got.Width != 32768 || got.Height != 32768 {
		t.Fatal("unexpected Prepare guard behavior")
	}
	t.Logf("%d raw bytes -> config %dx%d -> unchanged raw; no pixel decode attempted", len(raw), got.Width, got.Height)
}
