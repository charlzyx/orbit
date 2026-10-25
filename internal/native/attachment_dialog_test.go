package native

import "testing"

func TestAttachmentDialogSourceSingleSelection(t *testing.T) {
	opts := attachmentDialogOptions(nil, "/fixture")
	if opts.Multiple || opts.Directory || opts.DefaultPath != "/fixture" || len(opts.Filters) != 0 {
		t.Fatal("general file chooser must match single unfiltered Rust pick_file", opts)
	}
}
