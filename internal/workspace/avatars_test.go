package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAvatarIdentityAndOverrides(t *testing.T) {
	home := t.TempDir()
	requireNoError(t, os.MkdirAll(filepath.Join(home, ".orbit-pi"), 0700))
	requireNoError(t, os.WriteFile(filepath.Join(home, ".orbit-pi", "git-avatars.json"), []byte(`{"YOU@example.com":"override-user","invalid@example.com":"not a handle","wrong@example.com":42}`), 0600))
	overrides := LoadAvatarOverrides(home)
	if len(overrides) != 1 {
		t.Fatal(overrides)
	}
	cases := []struct{ name, email, want string }{{"Display Name", "you@EXAMPLE.com", "override-user"}, {"WrongHandle", "123+actual-user@users.noreply.github.com", "actual-user"}, {"Display Name", "legacy@users.noreply.github.com", "legacy"}, {"Display Name", "personal@example.com", ""}, {"good-handle", "personal@example.com", "good-handle"}, {"bad--handle", "personal@example.com", ""}}
	for _, c := range cases {
		if got := ResolveAvatarLogin(c.name, c.email, overrides); got != c.want {
			t.Fatal(c, got)
		}
	}
	if AvatarInitials("Ada Lovelace") != "AL" || AvatarInitials("git") != "GI" {
		t.Fatal("initials")
	}
	if AvatarHue("A", "x@example.com") != AvatarHue("B", "x@example.com") {
		t.Fatal("tint not stable by email")
	}
	if AvatarURL("valid", "", nil) != "https://github.com/valid.png?size=48" {
		t.Fatal("wrong endpoint")
	}
}
