package appbundle

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo"
)

// WorkspaceError retains the old selection until the user chooses a new folder.
type WorkspaceError struct {
	Home    string
	Layout  Layout
	Message string
}

func (e *WorkspaceError) Error() string { return e.Message }

// RuntimeError allows the user, not a silent fallback, to replace a missing
// saved runtime selection with the public runtime shipped by this app.
type RuntimeError struct{ Home, Message string }

func (e *RuntimeError) Error() string { return e.Message }

// ShowStartupError runs a minimal native lifecycle, not a shell/osascript
// workaround. Call only before the application's normal App.Run, on main.
func ShowStartupError(err error) error {
	mygo.App.SetName("GoPi")
	mygo.App.WhenReady(func() {
		var workspace *WorkspaceError
		var runtime *RuntimeError
		options := mygo.MessageOptions{Type: mygo.MessageError, Title: "GoPi could not start", Message: "GoPi could not start", Detail: err.Error(), Buttons: []string{"Quit"}}
		if errors.As(err, &workspace) {
			options.Buttons = []string{"Choose Folder…", "Quit"}
			options.CancelButton = 1
		} else if errors.As(err, &runtime) {
			options.Buttons = []string{"Use Bundled Pi", "Quit"}
			options.CancelButton = 1
		}
		result, dialogErr := mygo.Dialog.Message(options)
		if dialogErr != nil {
			mygo.App.Quit()
			return
		}
		if runtime != nil && result.Button == 0 {
			resetErr := preferences.New(runtime.Home).Update("runtime.json", func(value map[string]json.RawMessage) error { delete(value, "piBinary"); return nil })
			if resetErr == nil {
				mygo.App.Relaunch()
				return
			}
			_, _ = mygo.Dialog.Message(mygo.MessageOptions{Type: mygo.MessageError, Message: "The runtime selection could not be updated", Detail: resetErr.Error(), Buttons: []string{"Quit"}})
		}
		if workspace != nil && result.Button == 0 {
			paths, pickErr := mygo.Dialog.Open(mygo.OpenDialogOptions{Title: "Choose your GoPi workspace", Directory: true})
			if pickErr == nil && len(paths) > 0 {
				path := paths[0]
				info, statErr := os.Stat(path)
				if statErr == nil && info.IsDir() {
					pickErr = workspace.Layout.checkWritableLocation(path)
				} else {
					pickErr = errors.New("The chosen workspace is unavailable")
				}
				if pickErr == nil {
					pickErr = preferences.New(workspace.Home).SaveLastWorkspace(path)
				}
				if pickErr == nil {
					mygo.App.Relaunch()
					return
				}
			}
			if pickErr != nil {
				_, _ = mygo.Dialog.Message(mygo.MessageOptions{Type: mygo.MessageError, Message: "The workspace could not be selected", Detail: pickErr.Error(), Buttons: []string{"Quit"}})
			}
		}
		mygo.App.Quit()
	})
	return mygo.App.Run()
}
