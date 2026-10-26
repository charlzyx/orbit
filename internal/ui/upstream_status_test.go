package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
	"time"
)

func TestUpstreamStatusANSIPlainMounted(t *testing.T) {
	for _, tc := range []struct{ name, raw, want string }{
		{"sgr", "\x1b[31m红色 ready\x1b[0m", "红色 ready"},
		{"osc52_bel", "before\x1b]52;c;PRIVATE_PAYLOAD\aafter", "beforeafter"},
		{"osc52_st", "before\x1b]52;c;PRIVATE_PAYLOAD\x1b\\after", "beforeafter"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := fixture(t)
			a.SetSnapshot(Snapshot{ExtensionStatusText: tc.raw, ExtensionStatusAt: time.Now()})
			tt := ui.NewTester(a.View, 1470, 922)
			tt.Frame()
			if !tt.HasText(tc.want) || tt.HasText(tc.raw) {
				t.Fatalf("status retained controls: want=%q texts=%v", tc.want, tt.Texts())
			}
			if a.Model.ExtensionStatusText != tc.raw {
				t.Fatal("presentation mutated raw status")
			}
		})
	}
}

func TestUpstreamStatusANSIParserBoundaries(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{"plain π😀", "plain π😀"},
		{"a\x1b[2Jb", "ab"},
		{"a\x1b[31", "a"},
		{"a\x1b", "a"},
		{"a\x1b]unterminated", "a"},
		{"a\x1b_payload\x1b\\b", "ab"},
		{"a\x1bPpayload\ab", "ab"},
		{"a\x1b^payload\ab", "ab"},
		{"a\x1bXpayload\ab", "ab"},
		{"a\x1b(B", "aB"},
		{"tab\tline\nnext", "tab\tline\nnext"},
	} {
		if got := statusPlainText(tc.raw); got != tc.want {
			t.Errorf("%q got %q want %q", tc.raw, got, tc.want)
		}
	}
}
