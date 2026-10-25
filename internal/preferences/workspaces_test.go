package preferences

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestWorkspaceLegacySortAndSurgicalConcurrentEdits(t *testing.T) {
	s := New(t.TempDir())
	dir := t.TempDir()
	if e := s.Write("workspaces.json", map[string]any{"workspaces": []any{dir, map[string]any{"path": dir}}, "group_by_workspace": false, "sort": "unknown", "future": true}); e != nil {
		t.Fatal(e)
	}
	v := s.Workspaces()
	if len(v.Workspaces) != 1 || v.GroupBy != "one_list" || v.Sort != "last_updated" {
		t.Fatalf("%+v", v)
	}
	var wg sync.WaitGroup
	for n := 0; n < 20; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if e := s.AddWorkspace(filepath.Join(dir, fmt.Sprint(n))); e != nil {
				t.Error(e)
			}
		}(n)
	}
	wg.Wait()
	if len(s.Workspaces().Workspaces) != 21 {
		t.Fatal("concurrent add lost")
	}
	v = s.Workspaces()
	v.Sort = "manual"
	if e := s.SaveWorkspaces(v); e != nil {
		t.Fatal(e)
	}
	var raw map[string]json.RawMessage
	s.Read("workspaces.json", &raw)
	if string(raw["future"]) != "true" {
		t.Fatal("unknown field lost")
	}
	if e := s.RemoveWorkspace(dir); e != nil {
		t.Fatal(e)
	}
	if len(s.Workspaces().Workspaces) != 20 {
		t.Fatal("remove failed")
	}
	if e := s.SaveLastWorkspace(dir); e != nil {
		t.Fatal(e)
	}
	if s.LastWorkspace() != dir {
		t.Fatal("last workspace")
	}
	os.RemoveAll(dir)
	if s.LastWorkspace() != "" {
		t.Fatal("missing workspace restored")
	}
}
func TestWorkspaceMalformedWriteRefusal(t *testing.T) {
	s := New(t.TempDir())
	os.MkdirAll(s.RootDir, 0700)
	p := filepath.Join(s.RootDir, "workspaces.json")
	os.WriteFile(p, []byte("{bad"), 0600)
	if s.AddWorkspace(t.TempDir()) == nil {
		t.Fatal("malformed store overwritten")
	}
	b, _ := os.ReadFile(p)
	if string(b) != "{bad" {
		t.Fatal("data loss")
	}
}
