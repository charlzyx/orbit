package orbitui

import (
	"errors"
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestAttachmentPickerSingleResultAndCancellation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		paths []string
		err   error
		want  string
	}{
		{"unexpected multiple", []string{"first.txt", "second.txt"}, nil, "prefixfirst.txt "},
		{"cancel", nil, nil, "prefix"},
		{"error", []string{"first.txt"}, errors.New("fixture picker error"), "prefix"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			x := workspaceDialogFixture(t)
			a := x.UI
			tt := ui.NewTester(a.View, 1200, 800)
			if err := tt.Click("Message"); err != nil {
				t.Fatal(err)
			}
			tt.Type("prefix")
			release := holdFileDialog(t, x, Action{Kind: "BrowseAttachments"}, tc.paths, tc.err)
			release()
			drainControllerDialogUpdates(t, x, func() bool { return !x.fileDialogs.pending })
			tt.Frame()
			if a.Draft != tc.want {
				t.Fatal("picker result policy", a.Draft, tc.want)
			}
		})
	}
}
func TestAttachmentPickerChangeDoesNotRestrictMultiFileDrop(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.DropFiles(600, 400, "one.txt", "two.txt")
	tt.Frame()
	if a.Draft != "one.txt two.txt " {
		t.Fatal("drop lost multi-file behavior", a.Draft)
	}
}
