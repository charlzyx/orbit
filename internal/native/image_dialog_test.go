package native

import (
	"github.com/egoist/mygo"
	"reflect"
	"testing"
)

func TestImageDialogSourceSingleSelectionAndFilters(t *testing.T) {
	parent := new(mygo.Window)
	opts := imageDialogOptions(parent, "/fixture/current")
	if opts.Parent != parent || opts.DefaultPath != "/fixture/current" || opts.Title != "Attach image" || opts.Multiple || opts.Directory {
		t.Fatalf("wrong image options: %+v", opts)
	}
	want := []mygo.FileFilter{{Name: "Image", Extensions: []string{"png", "jpg", "jpeg", "webp", "gif", "bmp"}}}
	if !reflect.DeepEqual(opts.Filters, want) {
		t.Fatalf("source filter mismatch: %+v", opts.Filters)
	}
}
