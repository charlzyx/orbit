#!/usr/bin/env python3
"""Create an instrumentation-only Go overlay. Never rewrites runtime sources."""
from pathlib import Path
import argparse, hashlib, json, subprocess

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent

def sha(data): return hashlib.sha256(data).hexdigest()
def replace_once(source, old, new):
    if source.count(old) != 1:
        raise ValueError(f'Instrumentation anchor changed or ambiguous: {old!r}')
    return source.replace(old, new, 1)

def prepare(output):
    output.mkdir(parents=True, exist_ok=False)
    original = ROOT / 'internal/ui/run.go'
    source = original.read_text()
    replacements = [
        ('func (x *Controller) View(c *ui.Context) {', 'func (x *Controller) View(c *ui.Context) {\n\tworkspaceProbeViewEnter()\n\tdefer workspaceProbeViewLeave()'),
        ('func (x *Controller) SetWorkspace(path string) {', 'func (x *Controller) SetWorkspace(path string) {\n\tworkspaceProbeStage("set_workspace.begin", path)\n\tdefer workspaceProbeStage("set_workspace.return", path)'),
        ('\tx.changeWorkspaceServices(path)\n', '\tworkspaceProbeStage("set_workspace.services.before", path)\n\tx.changeWorkspaceServices(path)\n\tworkspaceProbeStage("set_workspace.services.after", path)\n'),
        ('\tx.UI.save(x.Agent.NewSession(path))', '\tworkspaceProbeStage("set_workspace.new_session.before", path)\n\tx.UI.save(x.Agent.NewSession(path))\n\tworkspaceProbeStage("set_workspace.new_session.after", path)'),
        ('\tcase "BrowseWorkspace":\n\t\tp, err := native.BrowseWorkspace(u.Window, x.workspace)', '\tcase "BrowseWorkspace":\n\t\tworkspaceProbeStage("browse.before", x.workspace)\n\t\tp, err := native.BrowseWorkspace(u.Window, x.workspace)\n\t\tworkspaceProbeStage("browse.return", map[string]any{"path": p, "error": fmt.Sprint(err)})'),
        ('\t\tw.OnReadyToShow(func() {', '\t\tw.OnReadyToShow(func() {\n\t\t\tworkspaceProbeStart(x, w)'),
    ]
    for old, new in replacements: source = replace_once(source, old, new)
    # fmt is already imported by the exact production run.go.
    if '\n\t"fmt"\n' not in source: raise ValueError('Expected existing fmt import')
    instrumented = output / 'run.go'
    instrumented.write_text(source)
    helper = output / 'workspace_probe_darwin.go'
    helper.write_bytes((HERE / 'workspace_probe_darwin.go.txt').read_bytes())
    virtual_helper = ROOT / 'internal/ui/workspace_probe_darwin.go'
    if virtual_helper.exists(): raise ValueError('Probe overlay would overwrite a source file')
    overlay = {'Replace': {str(original): str(instrumented), str(virtual_helper): str(helper)}}
    (output / 'overlay.json').write_text(json.dumps(overlay, indent=2) + '\n')
    inputs = {str(p.relative_to(ROOT)): sha(p.read_bytes()) for base in ('cmd/orbit', 'internal') for p in sorted((ROOT/base).rglob('*.go')) if not p.name.endswith('_test.go')}
    report = {'kind':'instrumented-production-source-overlay', 'source_commit': subprocess.check_output(['git','rev-parse','HEAD'], cwd=ROOT, text=True).strip(), 'source_status': subprocess.check_output(['git','status','--porcelain'],cwd=ROOT,text=True), 'production_source_sha256':inputs, 'original_run_go_sha256':sha(original.read_bytes()), 'instrumented_run_go_sha256':sha(instrumented.read_bytes()), 'probe_sha256':sha(helper.read_bytes()), 'limitations':['Not a byte-identical shipped executable. Production cmd/orbit main, Controller.View, Controller.Action, native.BrowseWorkspace, SetWorkspace and Pi lifecycle compiled from the recorded source; only stage logging, View-depth observation and an OnReadyToShow probe hook are added.', 'Probe posts AppKit mouse down/up through NSApplication event queue; it does not directly call Controller.Action or SetWorkspace.', 'NSOpenPanel selection uses setDirectoryURL: then real Ctrl-Return key down/up events through NSApplication, with an enabled actual default-button click fallback when available, cancellation uses cancel:. No stub dialog, TCC modification, accessibility authorization, external input automation or inference prompt.', 'Window screenshots use real native-content capture; panel bitmap capture may be transparent for remote views and is explicitly marked unavailable when so; an optional single-window screenshot runs only when screen recording access is already authorized.']}
    (output/'source-equivalence.json').write_text(json.dumps(report, indent=2)+'\n')
    print(output/'overlay.json')

if __name__ == '__main__':
    p=argparse.ArgumentParser(); p.add_argument('--output',type=Path,required=True); a=p.parse_args()
    if not a.output.is_absolute(): p.error('--output must be absolute')
    prepare(a.output)
