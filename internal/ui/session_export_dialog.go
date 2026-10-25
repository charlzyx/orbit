package orbitui

import "github.com/egoist/mygo"

func (x *Controller) exportSessionDialog() {
	owner := x.dialogTarget()
	target, err := x.Agent.CaptureSessionExportTarget()
	if err != nil {
		x.UI.save(err)
		return
	}
	x.saveDialog(owner, "Export session", "session.html", []mygo.FileFilter{{Name: "HTML", Extensions: []string{"html"}}}, func(path string, err error) {
		if !x.dialogTaskCurrent(owner) {
			return
		}
		if err != nil {
			x.UI.save(err)
			return
		}
		if path != "" {
			x.UI.save(x.Agent.ExportSessionFor(target, path))
		}
	})
}
