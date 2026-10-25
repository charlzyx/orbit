package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestVersionDoesNotCreateStateOrChangeEnvironment(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PI_BIN", "/sentinel")
	if code := run([]string{"--version"}); code != 0 {
		t.Fatal(code)
	}
	files, err := os.ReadDir(home)
	if err != nil || len(files) != 0 {
		t.Fatalf("version created state: %v %v", files, err)
	}
	if os.Getenv("PI_BIN") != "/sentinel" {
		t.Fatal("version changed env")
	}
}
func TestArgumentsKeepExplicitCLIOverrides(t *testing.T) {
	a, err := parseArguments([]string{"--home", "/private home", "--pi", "/selected pi", "--workspace", "/selected project", "--offline", "--session", "/session.jsonl"}, io.Discard)
	if err != nil || a.options.Home != "/private home" || a.options.PiBinary != "/selected pi" || a.options.Workspace != "/selected project" || !a.options.Offline || a.options.Session != "/session.jsonl" {
		t.Fatalf("%+v %v", a, err)
	}
	a, err = parseArguments([]string{filepath.Join("/", "positional workspace")}, io.Discard)
	if err != nil || a.options.Workspace != "/positional workspace" {
		t.Fatalf("%+v %v", a, err)
	}
}
