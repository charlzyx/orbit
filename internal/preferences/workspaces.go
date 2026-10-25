package preferences

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"time"
)

type Workspace struct {
	Path    string `json:"path"`
	AddedAt int64  `json:"added_at"`
	Icon    string `json:"icon,omitempty"`
	Tint    string `json:"tint,omitempty"`
}
type WorkspaceList struct {
	Workspaces     []Workspace `json:"workspaces"`
	Sort           string      `json:"sort"`
	GroupBy        string      `json:"group_by"`
	ArchivedFilter string      `json:"archived_filter"`
}

func CanonicalPath(path string) string {
	if v, e := filepath.EvalSymlinks(path); e == nil {
		return filepath.Clean(v)
	}
	if v, e := filepath.Abs(path); e == nil {
		return filepath.Clean(v)
	}
	return filepath.Clean(path)
}
func (s *Store) Workspaces() WorkspaceList {
	return parseWorkspaces(s.raw("workspaces.json"))
}
func parseWorkspaces(m map[string]json.RawMessage) WorkspaceList {
	v := WorkspaceList{Sort: str(m, "sort", "last_updated"), GroupBy: str(m, "group_by", "workspace"), ArchivedFilter: str(m, "archived_filter", "hide")}
	if !slices.Contains([]string{"manual", "alphabetical", "alphabetical_desc", "last_updated", "date_added", "session_count"}, v.Sort) {
		v.Sort = "last_updated"
	}
	if _, ok := m["group_by"]; !ok && !boolean(m, "group_by_workspace", true) {
		v.GroupBy = "one_list"
	}
	if v.GroupBy != "one_list" {
		v.GroupBy = "workspace"
	}
	if !slices.Contains([]string{"hide", "show", "only"}, v.ArchivedFilter) {
		v.ArchivedFilter = "hide"
	}
	var raw []json.RawMessage
	json.Unmarshal(m["workspaces"], &raw)
	seen := map[string]bool{}
	for _, r := range raw {
		var w Workspace
		var legacy string
		if json.Unmarshal(r, &legacy) == nil {
			w.Path = legacy
		} else if json.Unmarshal(r, &w) != nil {
			continue
		}
		if w.Path == "" {
			continue
		}
		w.Path = CanonicalPath(w.Path)
		if seen[w.Path] {
			continue
		}
		seen[w.Path] = true
		v.Workspaces = append(v.Workspaces, w)
	}
	return v
}
func (s *Store) SaveWorkspaces(v WorkspaceList) error {
	return s.Update("workspaces.json", func(m map[string]json.RawMessage) error {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		var fields map[string]json.RawMessage
		if err = json.Unmarshal(b, &fields); err != nil {
			return err
		}
		for k, raw := range fields {
			m[k] = raw
		}
		return nil
	})
}
func (s *Store) AddWorkspace(path string) error {
	path = CanonicalPath(path)
	return s.Update("workspaces.json", func(m map[string]json.RawMessage) error {
		v := parseWorkspaces(m)
		for _, w := range v.Workspaces {
			if w.Path == path {
				return nil
			}
		}
		v.Workspaces = append(v.Workspaces, Workspace{Path: path, AddedAt: time.Now().Unix()})
		raw, err := json.Marshal(v.Workspaces)
		if err == nil {
			m["workspaces"] = raw
		}
		return err
	})
}

// RemoveWorkspace only edits Orbit's list; Pi's session files are untouched.
func (s *Store) RemoveWorkspace(path string) error {
	path = CanonicalPath(path)
	return s.Update("workspaces.json", func(m map[string]json.RawMessage) error {
		v := parseWorkspaces(m)
		v.Workspaces = slices.DeleteFunc(v.Workspaces, func(w Workspace) bool { return w.Path == path })
		raw, err := json.Marshal(v.Workspaces)
		if err == nil {
			m["workspaces"] = raw
		}
		return err
	})
}
func (s *Store) LastWorkspace() string {
	p := str(s.raw("last-workspace.json"), "path", "")
	if p == "" {
		return ""
	}
	p = CanonicalPath(p)
	if st, e := os.Stat(p); e == nil && st.IsDir() {
		return p
	}
	return ""
}
func (s *Store) SaveLastWorkspace(path string) error {
	return s.Write("last-workspace.json", map[string]string{"path": CanonicalPath(path)})
}
