package native

import (
	"errors"
	"net/url"
	"strings"

	"github.com/egoist/mygo"
)

func BrowseWorkspace(parent *mygo.Window, current string) (string, error) {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: parent, Title: "Choose workspace", DefaultPath: current, Directory: true, CreateDirectories: true})
	if err != nil || len(paths) == 0 {
		return "", err
	}
	return paths[0], nil
}
func BrowseAttachments(parent *mygo.Window, current string) ([]string, error) {
	return mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: parent, Title: "Attach files", DefaultPath: current, Multiple: true})
}

// Mirrors Rust attach_image: one image, not the general multi-file chooser.
func imageDialogOptions(parent *mygo.Window, current string) mygo.OpenDialogOptions {
	return mygo.OpenDialogOptions{Parent: parent, Title: "Attach image", DefaultPath: current, Filters: []mygo.FileFilter{{Name: "Image", Extensions: []string{"png", "jpg", "jpeg", "webp", "gif", "bmp"}}}}
}
func BrowseImage(parent *mygo.Window, current string) ([]string, error) {
	return mygo.Dialog.Open(imageDialogOptions(parent, current))
}
func SaveFile(parent *mygo.Window, title, path string, filters []mygo.FileFilter) (string, error) {
	return mygo.Dialog.Save(mygo.SaveDialogOptions{Parent: parent, Title: title, DefaultPath: path, Filters: filters, CreateDirectories: true})
}
func Confirm(parent *mygo.Window, title, message, detail, confirm string) (bool, error) {
	r, err := mygo.Dialog.Message(mygo.MessageOptions{Parent: parent, Type: mygo.MessageWarning, Title: title, Message: message, Detail: detail, Buttons: []string{"Cancel", confirm}, DefaultButton: 0, CancelButton: 0})
	return err == nil && r.Button == 1, err
}
func TrashPath(path string) error { return mygo.Shell.TrashItem(path) }
func RevealPath(path string)      { mygo.Shell.ShowItemInFolder(path) }
func OpenPath(path string) error  { return mygo.Shell.OpenPath(path) }

// OpenURL is used for pi auth URLs, which must never launch arbitrary schemes.
func OpenURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || strings.ContainsAny(raw, "\x00\r\n") {
		return errors.New("only valid http(s) browser URLs may be opened")
	}
	return mygo.Shell.OpenExternal(raw)
}
