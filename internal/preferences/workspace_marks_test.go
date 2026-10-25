package preferences

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceMarkTargetedEdit(t *testing.T) {
	s := New(t.TempDir())
	root := t.TempDir()
	w := Workspace{Path: root, AddedAt: 10}
	other := Workspace{Path: t.TempDir(), AddedAt: 11, Icon: "star"}
	if err := s.SaveWorkspaces(WorkspaceList{Workspaces: []Workspace{w, other}, Sort: "manual", GroupBy: "one_list", ArchivedFilter: "show"}); err != nil {
		t.Fatal(err)
	}
	if err := s.EditWorkspace(w, s.Workspaces().Revision, false, "rocket-01", "blue"); err != nil {
		t.Fatal(err)
	}
	got := s.Workspaces()
	if got.Workspaces[0].Icon != "rocket-01" || got.Workspaces[0].Tint != "blue" || got.Workspaces[1] != other || got.Sort != "manual" || got.GroupBy != "one_list" || got.ArchivedFilter != "show" {
		t.Fatal(got)
	}
	for _, icon := range []string{"../../escape", "icons/star.svg"} {
		if err := s.EditWorkspace(w, s.Workspaces().Revision, false, icon, "blue"); err == nil {
			t.Fatal("unsafe icon accepted", icon)
		}
	}
	if err := s.EditWorkspace(w, s.Workspaces().Revision, false, "star", "madeup"); err == nil {
		t.Fatal("unknown tint accepted")
	}
	stale := w
	stale.AddedAt++
	if err := s.EditWorkspace(stale, s.Workspaces().Revision, true, "", ""); err == nil {
		t.Fatal("stale lifetime removed project")
	}
	path := filepath.Join(root, "saved.jsonl")
	if err := os.WriteFile(path, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.EditWorkspace(w, s.Workspaces().Revision, true, "", ""); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "keep" {
		t.Fatal("session modified", err)
	}
	if err := s.EditWorkspace(w, s.Workspaces().Revision, false, "star", ""); err == nil {
		t.Fatal("mark edit resurrected removed project")
	}
	if err := s.AddWorkspace(root); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceRevisionRejectsSameSecondABAFromOtherStore(t *testing.T) {
	home := t.TempDir()
	first, second := New(home), New(home)
	w := Workspace{Path: t.TempDir(), AddedAt: 123, Icon: "star"}
	if err := first.SaveWorkspaces(WorkspaceList{Workspaces: []Workspace{w}}); err != nil {
		t.Fatal(err)
	}
	old := first.Workspaces()
	if err := second.RemoveWorkspace(w.Path); err != nil {
		t.Fatal(err)
	}
	// Deliberately recreate the exact same seconds stamp and visible metadata.
	if err := second.SaveWorkspaces(WorkspaceList{Workspaces: []Workspace{w}}); err != nil {
		t.Fatal(err)
	}
	for _, remove := range []bool{false, true} {
		if err := first.EditWorkspace(w, old.Revision, remove, "rocket-01", "blue"); err == nil {
			t.Fatal("old snapshot edited re-added same-stamp workspace", remove)
		}
	}
	got := second.Workspaces()
	if len(got.Workspaces) != 1 || got.Workspaces[0] != w {
		t.Fatal("new lifetime was modified", got)
	}
	if err := first.EditWorkspace(w, got.Revision, false, "rocket-01", "blue"); err != nil {
		t.Fatal("fresh explicit retry blocked", err)
	}
}
func TestWorkspaceRevisionCoversWriteAndAliasStore(t *testing.T) {
	home := t.TempDir()
	s := New(home)
	w := Workspace{Path: t.TempDir(), AddedAt: 3}
	if err := s.SaveWorkspaces(WorkspaceList{Workspaces: []Workspace{w}}); err != nil {
		t.Fatal(err)
	}
	old := s.Workspaces()
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(home, alias); err != nil {
		t.Skip(err)
	}
	other := New(alias)
	if err := other.Write("workspaces.json", old); err != nil {
		t.Fatal(err)
	}
	if err := s.EditWorkspace(w, old.Revision, true, "", ""); err == nil {
		t.Fatal("direct Write through alias did not fence old edit")
	}
	before := s.Workspaces().Revision
	if err := other.AddWorkspace(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if s.Workspaces().Revision <= before {
		t.Fatal("AddWorkspace did not advance shared revision")
	}
}

func TestWorkspaceRevisionCaseAliasWhenSupported(t *testing.T) {
	home := filepath.Join(t.TempDir(), "CaseHome")
	if err := os.MkdirAll(home, 0700); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(filepath.Dir(home), strings.ToLower(filepath.Base(home)))
	a, err := os.Stat(home)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.Stat(alias)
	if err != nil || !os.SameFile(a, b) {
		t.Skip("case-sensitive filesystem; native Mac coverage pending")
	}
	first, second := New(home), New(alias)
	w := Workspace{Path: t.TempDir(), AddedAt: 3}
	if err := first.SaveWorkspaces(WorkspaceList{Workspaces: []Workspace{w}}); err != nil {
		t.Fatal(err)
	}
	old := first.Workspaces()
	if err := second.Write("workspaces.json", old); err != nil {
		t.Fatal(err)
	}
	if err := first.EditWorkspace(w, old.Revision, true, "", ""); err == nil {
		t.Fatal("case alias bypassed revision")
	}
}
