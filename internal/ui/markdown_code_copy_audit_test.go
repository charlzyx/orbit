package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"strings"
	"testing"
)

func TestMountedCodeCopyMatchesRustPayload(t *testing.T) {
	cases := []struct {
		name, source, want string
		stream             bool
	}{
		{"closed", "```sh\nprintf x\n```", "printf x", false},
		{"intentional final blank", "```sh\nprintf x\n\n```", "printf x\n", false},
		{"CRLF", "```sh\r\nprintf x\r\n```", "printf x", false},
		{"multiple final blanks", "```sh\nprintf x\n\n\n```", "printf x\n\n", false},
		{"tabs unicode", "```text\n\t用户 α  \n```", "\t用户 α  ", false},
		{"empty", "```sh\n```", "", false},
		{"open live", "```sh\nprintf x", "printf x", true},
		{"open settled", "```sh\nprintf x", "printf x", false},
		{"whitespace", "```text\n  spaced  \n\n```", "  spaced  \n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := fixture(t)
			a.SetSnapshot(Snapshot{SessionID: "code-copy", Messages: []Message{{ID: "code", Role: "assistant", Text: tc.source, Streaming: tc.stream}}})
			tt := ui.NewTester(a.View, 1470, 922)
			tt.SetClipboard("not copied")
			if err := tt.Click("Copy code"); err != nil {
				t.Fatal("actual code copy unavailable", err)
			}
			if got := tt.Clipboard(); got != tc.want {
				t.Fatalf("clipboard=%q want Rust lines.join=%q", got, tc.want)
			}
		})
	}
}

func TestMountedCodeCopyOwnerAndChangedSource(t *testing.T) {
	for _, change := range []string{"content", "runtime", "aba", "close"} {
		t.Run(change, func(t *testing.T) {
			a := fixture(t)
			snap := Snapshot{RuntimeID: "runtime-a", SessionID: "same", Workspace: "/work/a", Messages: []Message{{ID: "code", Role: "assistant", Text: "```sh\nold value\n```"}}}
			a.SetSnapshot(snap)
			tt := ui.NewTester(a.View, 1470, 922)
			rect, ok := tt.Find("Copy code")
			if !ok {
				t.Fatal("copy missing")
			}
			tt.SetClipboard("sentinel")
			tt.Press(rect.X+rect.W/2, rect.Y+rect.H/2)
			switch change {
			case "content":
				snap.Messages[0].Text = "```sh\nnew value\n```"
				a.SetSnapshot(snap)
			case "runtime":
				snap.RuntimeID = "runtime-b"
				a.SetSnapshot(snap)
			case "aba":
				snap.RuntimeID = "runtime-b"
				a.SetSnapshot(snap)
				snap.RuntimeID = "runtime-a"
				a.SetSnapshot(snap)
			case "close":
				a.closeInlineImages()
			}
			tt.Frame()
			tt.Release(rect.X+rect.W/2, rect.Y+rect.H/2)
			tt.Frame()
			if got := tt.Clipboard(); got != "sentinel" {
				t.Fatalf("stale press copied %q", got)
			}
			if change != "close" {
				if err := tt.Click("Copy code"); err != nil {
					t.Fatal(err)
				}
				want := "old value"
				if change == "content" {
					want = "new value"
				}
				if tt.Clipboard() != want {
					t.Fatal(tt.Clipboard())
				}
			}
		})
	}
}

func TestMountedCodeCopyStreamAppendSettleAndAlert(t *testing.T) {
	a := fixture(t)
	snap := Snapshot{SessionID: "stream-code", Messages: []Message{{ID: "code", Role: "assistant", Text: "```sh\none", Streaming: true}}}
	a.SetSnapshot(snap)
	tt := ui.NewTester(a.View, 1470, 922)
	for _, text := range []string{"```sh\none", "```sh\none\ntwo", "```sh\none\ntwo\n```"} {
		snap.Messages[0].Text = text
		a.SetSnapshot(snap)
		tt.Frame()
		if err := tt.Click("Copy code"); err != nil {
			t.Fatal(err)
		}
		want := "one\ntwo"
		if text == "```sh\none" {
			want = "one"
		}
		if tt.Clipboard() != want {
			t.Fatal(tt.Clipboard())
		}
	}
	snap.Messages[0].Streaming = false
	snap.Messages[0].Text = "> [!NOTE]\n> ```sh\n> alert code\n> ```"
	a.SetSnapshot(snap)
	tt.Frame()
	if err := tt.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != "alert code" {
		t.Fatal(tt.Clipboard())
	}
}

func TestCodeCopyIdleDoesNotReadPayloadAndKeyboardWorks(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{SessionID: "idle-code"})
	a.md.Theme = a.Theme.Markdown()
	reads := 0
	tt := ui.NewTester(func(c *ui.Context) {
		a.codeCopyHeader(c, "sh", a.inlineImageToken(), 1, func() bool { return true }, func() string { reads++; return "payload\n" })
	}, 600, 160)
	for i := 0; i < 30; i++ {
		tt.Frame()
	}
	if reads != 0 {
		t.Fatal("idle body materialization", reads)
	}
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnter)
	if reads != 1 || tt.Clipboard() != "payload" {
		t.Fatalf("keyboard copy reads=%d clipboard=%q", reads, tt.Clipboard())
	}
	for i := 0; i < 30; i++ {
		tt.Frame()
	}
	if reads != 1 {
		t.Fatal("feedback frames reread payload", reads)
	}
}

func TestMountedCodeCopyPreservesCrossMessageSelection(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{SessionID: "selection-code", Messages: []Message{{ID: "before", Role: "user", Text: "BEFORE CODE"}, {ID: "code", Role: "assistant", Text: "```text\nSELECTABLE CODE\n```"}, {ID: "after", Role: "user", Text: "AFTER CODE"}}})
	tt := ui.NewTester(a.View, 1470, 922)
	first, ok := tt.Find("BEFORE CODE")
	if !ok {
		t.Fatal("before missing")
	}
	last, ok := tt.Find("AFTER CODE")
	if !ok {
		t.Fatal("after missing")
	}
	tt.Press(first.X, first.Y+first.H/2)
	tt.Move(last.X+last.W, last.Y+last.H/2)
	tt.Release(last.X+last.W, last.Y+last.H/2)
	tt.Key(ui.Cmd, ui.KeyC)
	for _, want := range []string{"BEFORE CODE", "SELECTABLE CODE", "AFTER CODE"} {
		if !strings.Contains(tt.Clipboard(), want) {
			t.Fatal("selection missing", want, tt.Clipboard())
		}
	}
}

func TestCodeCopyTwoBlocksKeyboardAndHorizontalScroll(t *testing.T) {
	a := fixture(t)
	a.md.Theme = a.Theme.Markdown()
	a.SetSnapshot(Snapshot{SessionID: "two-code"})
	first := "```sh\nfirst block\n```"
	second := "```sh\nsecond block\n```"
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Children(func() { a.chatMarkdown(c.Key("first"), first, false); a.chatMarkdown(c.Key("second"), second, false) })
	}, 600, 400)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnter)
	if tt.Clipboard() != "first block" {
		t.Fatal("first keyboard activation", tt.Clipboard())
	}
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnter)
	if tt.Clipboard() != "second block" {
		t.Fatal("second keyboard activation", tt.Clipboard())
	}
	// A source revision in the first block does not replace the second payload.
	first = "```sh\nfirst changed\n```"
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if tt.Clipboard() != "second block" {
		t.Fatal("neighbor revision corrupted focused block", tt.Clipboard())
	}
	line := strings.Repeat("long code ", 40)
	scroll := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Width(300).Selectable().Children(func() { a.chatMarkdown(c, lineFence(line), false) })
	}, 400, 240)
	before, ok := scroll.Find(line)
	if !ok {
		t.Fatal("code body missing")
	}
	scroll.Scroll(before.X+10, before.Y+before.H/2, 300, 0)
	scroll.Frame()
	after, ok := scroll.Find(line)
	if !ok || after.X >= before.X {
		t.Fatalf("body horizontal scroll lost: %v %v", before, after)
	}
}
func lineFence(line string) string { return "```text\n" + line + "\n```" }

func TestMountedCodeCopyLogicalRPCProjection(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{SessionID: "logical-code", RawMessages: []agent.Message{{ID: "u", Role: "user", Text: "question"}, {ID: "a", Role: "assistant", Text: "```sh\nlogical code\n```"}}})
	tt := ui.NewTester(a.View, 1470, 922)
	if err := tt.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != "logical code" {
		t.Fatal(tt.Clipboard())
	}
}
