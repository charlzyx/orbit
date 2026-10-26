package preferences

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCommitModelPrivateRoundtripAndFailures(t *testing.T) {
	s := New(t.TempDir())
	if m, e := s.CommitModel(); e != nil || m != nil {
		t.Fatal(m, e)
	}
	want := &Model{Provider: "provider", ID: "id"}
	if e := s.SaveCommitModel(want); e != nil {
		t.Fatal(e)
	}
	restarted := New(s.Home)
	if got, e := restarted.CommitModel(); e != nil || !reflect.DeepEqual(got, want) {
		t.Fatal(got, e)
	}
	stat, e := os.Stat(filepath.Join(s.RootDir, "commit-model.json"))
	if e != nil || stat.Mode().Perm() != 0600 {
		t.Fatal(stat, e)
	}
	for _, raw := range []string{`{"model":null}`, `{}`, `{"model":{"id":"id"}}`, `{"model":{"provider":"p","id":""}}`} {
		if e = os.WriteFile(filepath.Join(s.RootDir, "commit-model.json"), []byte(raw), 0600); e != nil {
			t.Fatal(e)
		}
		if got, e := s.CommitModel(); e != nil || got != nil {
			t.Fatal(raw, got, e)
		}
	}
	for _, raw := range []string{`{bad`, `{"model":{"provider":7,"id":"id"}}`} {
		_ = os.WriteFile(filepath.Join(s.RootDir, "commit-model.json"), []byte(raw), 0600)
		if got, e := s.CommitModel(); e == nil || got != nil {
			t.Fatal(raw, got, e)
		}
	}
	if e = s.SaveCommitModel(nil); e != nil {
		t.Fatal(e)
	}
	raw, _ := os.ReadFile(filepath.Join(s.RootDir, "commit-model.json"))
	if string(raw) != "{\n  \"model\": null\n}\n" {
		t.Fatal(string(raw))
	}
	if e = os.Remove(filepath.Join(s.RootDir, "commit-model.json")); e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(filepath.Join(s.RootDir, "commit-model.json"), 0700); e != nil {
		t.Fatal(e)
	}
	if e = s.SaveCommitModel(want); e == nil {
		t.Fatal("directory write succeeded")
	}
}
