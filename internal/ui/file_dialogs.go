package orbitui

import (
	"bytes"
	"errors"
	"os"
	"strings"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo"
)

type nativeFileDialogState struct {
	pending     bool
	background  func(mygo.OpenDialogOptions) ([]string, error)
	attachments func(*mygo.Window, string) ([]string, error)
	save        func(*mygo.Window, string, string, []mygo.FileFilter) (string, error)
}

type nativeDialogTarget struct {
	parent    *mygo.Window
	epoch     uint64
	workspace string
	// Existing opaque key+generation identifies empty-ID drafts too, without
	// treating normal get_state session-ID hydration as a different task.
	task agent.WorkflowTarget
}

func (x *Controller) dialogTarget() nativeDialogTarget {
	return nativeDialogTarget{x.UI.Window, x.workspaceBrowse.epoch, x.workspace, x.Agent.CaptureWorkflowTarget()}
}
func (x *Controller) dialogAlive(target nativeDialogTarget) bool {
	select {
	case <-x.stop:
		return false
	default:
	}
	return target.parent == x.UI.Window && (target.parent == nil || !target.parent.IsDestroyed())
}
func (x *Controller) dialogTaskCurrent(target nativeDialogTarget) bool {
	if !x.dialogAlive(target) || target.epoch != x.workspaceBrowse.epoch || target.workspace != x.workspace {
		return false
	}
	return x.Agent.CaptureWorkflowTarget() == target.task
}

// Native waits happen off the immediate-mode view. Completion/ownership checks
// run on the UI queue; user-confirmed side effects only start after those checks.
func (x *Controller) fileDialog(target nativeDialogTarget, open func() ([]string, error), apply func([]string, error)) {
	if x.fileDialogs.pending || x.workspaceBrowse.pending || !x.dialogAlive(target) {
		return
	}
	x.fileDialogs.pending = true
	go func() {
		select {
		case <-x.stop:
			return
		default:
		}
		paths, err := open()
		x.enqueue(func() {
			x.fileDialogs.pending = false
			if !x.dialogAlive(target) {
				return
			}
			apply(paths, err)
		})
	}()
}
func (x *Controller) browseAttachmentsDialog() {
	target := x.dialogTarget()
	open := x.fileDialogs.attachments
	if open == nil {
		open = native.BrowseAttachments
	}
	x.fileDialog(target, func() ([]string, error) { return open(target.parent, target.workspace) }, func(paths []string, err error) {
		if !x.dialogTaskCurrent(target) {
			return
		}
		if err != nil {
			x.UI.save(err)
			return
		}
		x.UI.AddPaths(paths)
	})
}
func (x *Controller) saveDialog(target nativeDialogTarget, title, path string, filters []mygo.FileFilter, apply func(string, error)) {
	save := x.fileDialogs.save
	if save == nil {
		save = native.SaveFile
	}
	x.fileDialog(target, func() ([]string, error) { p, e := save(target.parent, title, path, filters); return []string{p}, e }, func(paths []string, err error) {
		p := ""
		if len(paths) > 0 {
			p = paths[0]
		}
		apply(p, err)
	})
}
func cloneUsageExport(index workbench.UsageIndex, filter workbench.UsageFilter, options workbench.UsageExportOptions) (workbench.UsageIndex, workbench.UsageFilter, workbench.UsageExportOptions) {
	index.ModelCatalog = append([]workbench.UsageModelRef(nil), index.ModelCatalog...)
	index.Sessions = append([]workbench.UsageSession(nil), index.Sessions...)
	index.Errors = append([]workbench.UsageError(nil), index.Errors...)
	index.Records = append([]workbench.UsageRecord(nil), index.Records...)
	index.Tools = append([]workbench.ToolRun(nil), index.Tools...)
	index.Turns = append([]workbench.UsageTurn(nil), index.Turns...)
	index.Warnings = append([]string(nil), index.Warnings...)
	filter = cloneUsageFilter(filter)
	options.Insights = append([]string(nil), options.Insights...)
	titles := make(map[string]string, len(options.SessionTitles))
	for k, v := range options.SessionTitles {
		titles[k] = v
	}
	options.SessionTitles = titles
	return index, filter, options
}
func (x *Controller) exportUsageDialog(action Action) {
	index, ok := action.Data["index"].(workbench.UsageIndex)
	if !ok {
		x.UI.save(errors.New("Usage data is not ready to export"))
		return
	}
	filter, _ := action.Data["filter"].(workbench.UsageFilter)
	options, _ := action.Data["options"].(workbench.UsageExportOptions)
	index, filter, options = cloneUsageExport(index, filter, options)
	format := action.Value
	if format != "json" {
		format = "csv"
	}
	target := x.dialogTarget()
	// Rust usage/page.rs::export captures the view/index before awaiting the
	// panel. Later filter/page/workspace changes must not change the chosen data.
	x.saveDialog(target, "Export usage", "orbit-usage."+format, []mygo.FileFilter{{Name: strings.ToUpper(format), Extensions: []string{format}}}, func(path string, err error) {
		if err != nil {
			x.UI.save(err)
			return
		}
		if path == "" {
			return
		}
		go func() {
			var body bytes.Buffer
			if format == "json" {
				err = index.ExportJSON(&body, filter, options)
			} else {
				err = index.ExportCSV(&body, filter)
			}
			// Serialize before opening/truncating a destination. Native SaveFile still
			// owns the ordinary overwrite confirmation; no guessed path is written.
			if err == nil {
				err = os.WriteFile(path, body.Bytes(), 0666)
			}
			x.enqueue(func() {
				if x.dialogAlive(target) {
					x.UI.save(err)
				}
			})
		}()
	})
}
