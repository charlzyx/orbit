package native

import (
	"io"
	"strings"
	"testing"
)

type outputReader struct{ io.Reader }

func TestRuntimeOutputReaderFromCannotBypassLimit(t *testing.T) {
	b := boundedOutput{limit: 16}
	n, err := io.Copy(&b, outputReader{strings.NewReader(strings.Repeat("a", 100))})
	if err != nil || n != 100 || b.Len() != 16 {
		t.Fatalf("copy=%d stored=%d err=%v", n, b.Len(), err)
	}
	if n, err := b.Write([]byte("more")); n != 4 || err != nil || b.Len() != 16 {
		t.Fatal(n, err, b.Len())
	}
}
