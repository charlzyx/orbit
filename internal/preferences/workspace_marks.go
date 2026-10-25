package preferences

import (
	"encoding/json"
	"errors"
	"slices"
)

// Curated source: orbit-pi/src/workspace_mark.rs. Never accept asset paths.
var WorkspaceIcons = []string{"folder", "folder-open", "folder-sync", "folder-archive", "branch", "rocket-01", "star", "spark", "cloud", "terminal", "monitor", "server-stack", "git-fork", "github", "image", "lock", "clock", "tag-01", "task", "compose", "magic-wand", "start-up", "extensions", "chat", "archive", "contrast", "side-by-side"}
var WorkspaceTints = []string{"default", "accent", "green", "red", "yellow", "blue", "purple", "pink"}

var ErrWorkspaceListChanged = errors.New("workspace list changed; reopen its menu and retry")

// EditWorkspace changes only the captured project, never the current project.
// Called on a worker: canonicalization and store I/O must not run in render.
func (s *Store) EditWorkspace(expected Workspace, revision uint64, remove bool, icon, tint string) error {
	if icon != "" && !slices.Contains(WorkspaceIcons, icon) {
		return errors.New("unknown workspace icon")
	}
	if tint != "" && !slices.Contains(WorkspaceTints, tint) {
		return errors.New("unknown workspace tint")
	}
	path := CanonicalPath(expected.Path)
	return s.Update("workspaces.json", func(m map[string]json.RawMessage) error {
		if workspaceFiles.revision[s.workspaceFileKey()] != revision {
			return ErrWorkspaceListChanged
		}
		v := parseWorkspaces(m)
		i := slices.IndexFunc(v.Workspaces, func(w Workspace) bool { return w.Path == path && w.AddedAt == expected.AddedAt })
		if i < 0 {
			return ErrWorkspaceListChanged
		}
		if remove {
			v.Workspaces = slices.Delete(v.Workspaces, i, i+1)
		} else {
			v.Workspaces[i].Icon, v.Workspaces[i].Tint = icon, tint
		}
		raw, err := json.Marshal(v.Workspaces)
		if err == nil {
			m["workspaces"] = raw
		}
		return err
	})
}
