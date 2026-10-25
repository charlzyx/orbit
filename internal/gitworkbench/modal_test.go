package gitworkbench

import "testing"

func TestHasModalTracksDialogAndConfirmation(t *testing.T) {
	var absent *State
	if absent.HasModal() {
		t.Fatal("nil page is not modal")
	}
	s := &State{}
	for _, test := range []struct{ dialog, confirm, want bool }{{false, false, false}, {true, false, true}, {false, true, true}, {true, true, true}, {false, false, false}} {
		s.dialogOpen, s.confirmOpen = test.dialog, test.confirm
		if s.HasModal() != test.want {
			t.Fatalf("dialog=%v confirm=%v", test.dialog, test.confirm)
		}
	}
}
