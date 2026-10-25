package preferences

import (
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestPiConfigPathsAndPrecedence(t *testing.T) {
	home, private := t.TempDir(), t.TempDir()
	cases := []struct{ input, want string }{{"", filepath.Join(home, ".pi", "agent")}, {"~", home}, {"~/custom agent", filepath.Join(home, "custom agent")}, {filepath.Join(home, "absolute"), filepath.Join(home, "absolute")}}
	for _, tt := range cases {
		if got := ResolvePiDir(home, tt.input); got != tt.want {
			t.Errorf("%q resolved as %q, want %q", tt.input, got, tt.want)
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if got := ResolvePiDir(home, "relative/config"); got != filepath.Join(cwd, "relative", "config") {
		t.Fatal(got)
	}
	if got := ResolvePiDir(home, "~someone/config"); got != filepath.Join(cwd, "~someone", "config") {
		t.Fatal("expanded named user's home")
	}
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("PI_CODING_AGENT_DIR", "~/custom agent")
	normal, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	if normal.PiDir != filepath.Join(home, "custom agent") {
		t.Fatal("environment override lost")
	}
	isolated := New(private)
	if isolated.UserHome != private || isolated.PiDir != filepath.Join(private, ".pi", "agent") {
		t.Fatal("explicit home borrowed ambient config")
	}
}
func TestPrivateOrbitPreferencesAndSharedPiEnvironment(t *testing.T) {
	private, user, pi := t.TempDir(), t.TempDir(), t.TempDir()
	s := NewWithPi(private, user, pi)
	if err := s.SaveAccess("supervised"); err != nil {
		t.Fatal(err)
	}
	if NewWithPi(private, user, pi).Access() != "supervised" {
		t.Fatal("private preferences did not persist")
	}
	for _, root := range []string{user, pi} {
		if _, err := os.Stat(filepath.Join(root, ".orbit-pi")); !os.IsNotExist(err) {
			t.Fatal("wrote preferences into Pi home")
		}
	}
	env := s.PiEnvironment([]string{"HOME=wrong", "USERPROFILE=wrong", "PI_CODING_AGENT_DIR=wrong", "ORBIT_PREFERENCES_HOME=wrong", "SYNTHETIC_API_KEY=fixture", "PATH=/synthetic/bin"})
	got := map[string]string{}
	for _, entry := range env {
		k, v, _ := strings.Cut(entry, "=")
		if _, exists := got[k]; exists {
			t.Fatal("duplicate variable", k)
		}
		got[k] = v
	}
	want := s.PiEnvOverrides()
	want["SYNTHETIC_API_KEY"] = "fixture"
	want["PATH"] = "/synthetic/bin"
	if !reflect.DeepEqual(got, want) {
		t.Fatal("helper environment differs from main child configuration")
	}
}

func TestMalformedPiFileURLsNeverRedirectIntoAnotherTree(t *testing.T) {
	home := t.TempDir()
	invalid := []string{"file:///bad%2Fencoded-separator", "file:///bad%escape"}
	if runtime.GOOS != "windows" {
		invalid = append(invalid, "file://remote-host/config")
	}
	for _, input := range invalid {
		if ValidatePiDir(input) == nil {
			t.Fatal("accepted malformed Pi config URL")
		}
		if got := ResolvePiDir(home, input); got != input {
			t.Fatal("invalid config silently redirected to another tree")
		}
	}
	if runtime.GOOS != "windows" {
		for input, want := range map[string]string{"file://LOCALHOST/tmp/config": "/tmp/config", "file://localhost": "/", "file://": "/"} {
			if err := ValidatePiDir(input); err != nil {
				t.Fatal(err)
			}
			if got := ResolvePiDir(home, input); got != want {
				t.Fatal("valid file URL changed identity", got)
			}
		}
	}
	valid := (&url.URL{Scheme: "file", Path: filepath.ToSlash(filepath.Join(home, "config with spaces"))}).String()
	if runtime.GOOS == "windows" {
		valid = "file:///" + filepath.ToSlash(filepath.Join(home, "config with spaces"))
	}
	if err := ValidatePiDir(valid); err != nil {
		t.Fatal(err)
	}
	if got := ResolvePiDir(home, valid); got != filepath.Join(home, "config with spaces") {
		t.Fatal("valid file URL not resolved", got)
	}
}
