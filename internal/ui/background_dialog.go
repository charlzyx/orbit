package orbitui

import "github.com/egoist/mygo"

// Rust settings.rs::background_choose uses AsyncFileDialog. Background choice
// is an application preference, not a task attachment: navigation/task changes
// do not invalidate it, but a newer background operation or closed owner does.
func (x *Controller) browseBackgroundDialog() {
	s := &x.UI.backdrop
	if s.loading || s.closed {
		return
	}
	target, generation := x.dialogTarget(), s.generation
	options := mygo.OpenDialogOptions{Parent: target.parent, Title: x.UI.text("settings.choose_background_title", "Choose a background image"), Filters: []mygo.FileFilter{{Name: "Images", Extensions: []string{"png", "jpg", "jpeg", "gif", "webp", "bmp", "tiff"}}}}
	open := x.fileDialogs.background
	if open == nil {
		open = mygo.Dialog.Open
	}
	x.fileDialog(target, func() ([]string, error) { return open(options) }, func(paths []string, err error) {
		if s.closed || s.generation != generation {
			return
		}
		if err != nil {
			x.UI.save(err)
		} else if len(paths) > 0 {
			x.UI.chooseBackground(paths[0])
		}
	})
}
