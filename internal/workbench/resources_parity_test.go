package workbench

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSkillDiscoveryLinksStopsAtRootAndSkipsHidden(t *testing.T) {
	s := testService(t)
	root := filepath.Join(s.Workspace, ".pi", "skills")
	put(t, filepath.Join(root, "parent", "SKILL.md"), "---\r\nname: parent\r\ndescription: Parent\r\n---\r\nBody\r\n")
	put(t, filepath.Join(root, "parent", "nested", "SKILL.md"), "nested")
	for _, name := range []string{".hidden", "node_modules"} {
		put(t, filepath.Join(root, name, "SKILL.md"), name)
	}
	target := filepath.Join(s.Preferences.Home, "external")
	put(t, filepath.Join(target, "SKILL.md"), "---\nname: linked\ndescription: Link\n---\nLinked body.")
	if e := os.Symlink(target, filepath.Join(root, "linked")); e != nil {
		t.Skip(e)
	}
	if e := os.Symlink(root, filepath.Join(root, "cycle")); e != nil {
		t.Fatal(e)
	}
	skills, errs := s.Skills()
	if len(errs) > 0 || len(skills) != 2 {
		t.Fatal(skills, errs)
	}
	if skills[1].Name != "parent" || skills[1].Body() != "Body\n" || !skills[1].Valid() || skills[0].InvokeCommand() != "/skill:linked" {
		t.Fatal(skills)
	}
	// Removing a directory symlink removes the link, not the externally stored skill.
	if e := s.DeleteSkill(skills[0].Path); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(filepath.Join(target, "SKILL.md")); e != nil {
		t.Fatal("followed symlink during removal", e)
	}
}
func TestSkillRootDeleteAndOverrideSafety(t *testing.T) {
	s := testService(t)
	root := filepath.Join(s.Workspace, ".pi", "skills")
	path := filepath.Join(root, "SKILL.md")
	put(t, path, "Root")
	put(t, filepath.Join(root, "keep.txt"), "keep")
	settings := filepath.Join(s.Workspace, ".pi", "settings.json")
	put(t, settings, fmt.Sprintf(`{"unknown":42,"skills":["-%s","+%s","custom",{"path":"-%s","other":true}]}`, path, root, path))
	skills, _ := s.Skills()
	if len(skills) != 1 || skills[0].Enabled {
		t.Fatal(skills)
	}
	if e := s.SetSkillEnabled(skills[0], true); e != nil {
		t.Fatal(e)
	}
	m, _ := readObject(settings)
	entries := m["skills"].([]any)
	if len(entries) != 2 || entries[0] != "custom" || m["unknown"] != float64(42) {
		t.Fatal(m)
	}
	// Object overrides are read but preserved by the source string-only write contract.
	skills, _ = s.Skills()
	if skills[0].Enabled {
		t.Fatal("object override accidentally removed")
	}
	if e := s.DeleteSkill(path); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(filepath.Join(root, "keep.txt")); e != nil {
		t.Fatal("deleted discovery root")
	}
	put(t, path, "Root")
	put(t, settings, `{"skills":{"legacy":true}}`)
	skills, _ = s.Skills()
	before, _ := os.ReadFile(settings)
	if e := s.SetSkillEnabled(skills[0], false); e == nil {
		t.Fatal("clobbered old schema")
	}
	after, _ := os.ReadFile(settings)
	if string(before) != string(after) {
		t.Fatal("changed old schema")
	}
}
func TestSkillEnableNoopAndHomeScope(t *testing.T) {
	s := testService(t)
	path := filepath.Join(s.Preferences.Home, ".agents", "skills", "global", "SKILL.md")
	put(t, path, "---\nname: home\ndescription: test\n---\n")
	skills, errs := s.Skills()
	if len(errs) > 0 || len(skills) != 1 || skills[0].Scope != "user" {
		t.Fatal(skills, errs)
	}
	if e := s.SetSkillEnabled(skills[0], true); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(filepath.Join(s.Preferences.PiDir, "settings.json")); !os.IsNotExist(e) {
		t.Fatal("no-op enable created a settings file")
	}
}
func TestPluginProbeSkipsPinnedAndPreservesSkipStore(t *testing.T) {
	s := testService(t)
	pkgs := []Plugin{{Source: "npm:pkg", Name: "pkg", Kind: "npm", Scope: "global", Installed: true, Version: "1.2"}, {Source: "npm:pin@1.0", Name: "pin", Kind: "npm", Installed: true, Version: "1.0"}, {Source: "github:a/b#pinned", Name: "git", Kind: "git", Installed: true}, {Source: "./local", Name: "local", Kind: "local", Installed: true}}
	calls := 0
	probe := func(ctx context.Context, dir, binary string, args ...string) (string, error) {
		calls++
		if binary != "npm" && runtime.GOOS != "windows" {
			t.Fatal(binary)
		}
		return `"1.2.1"`, nil
	}
	got, errs := s.findPluginUpdates(context.Background(), pkgs, nil, probe)
	if len(errs) > 0 || len(got) != 1 || calls != 1 || got[0].Latest != "1.2.1" {
		t.Fatal(got, errs, calls)
	}
	if e := s.SkipPluginUpdate("npm:pkg", "1.2.1"); e != nil {
		t.Fatal(e)
	}
	skips, e := s.SkippedPluginUpdates()
	if e != nil {
		t.Fatal(e)
	}
	got, errs = s.findPluginUpdates(context.Background(), pkgs, skips, probe)
	if len(got) != 0 || len(errs) > 0 {
		t.Fatal(got, errs)
	}
	if e := s.SkipPluginUpdate("npm:pkg", ""); e != nil {
		t.Fatal(e)
	}
	skips, _ = s.SkippedPluginUpdates()
	if len(skips) != 0 {
		t.Fatal(skips)
	}
	for _, source := range []string{"npm:pkg@latest", "npm:pkg;bad", "npm:-flag", "npm:pkg name", "npm:@scope/pkg@2"} {
		if npmUpdateName(source) != "" {
			t.Fatal(source)
		}
	}
	for _, out := range []string{`"1.2.3"`, `["1.2.3"]`, `{"version":"1.2.3"}`, "1.2.3\n"} {
		if npmVersion(out) != "1.2.3" {
			t.Fatal(out)
		}
	}
	if newerPluginVersion("1.2.0+build.7", "1.2") || !newerPluginVersion("1.10", "1.9") {
		t.Fatal("version semantics differ")
	}
}
func TestPluginBoundedParallelErrorsAndSkipMutation(t *testing.T) {
	s := testService(t)
	var pkgs []Plugin
	for n := range 15 {
		pkgs = append(pkgs, Plugin{Source: fmt.Sprintf("npm:pkg%d", n), Name: fmt.Sprint(n), Kind: "npm", Installed: true, Version: "1"})
	}
	var active, peak atomic.Int32
	probe := func(ctx context.Context, dir, binary string, args ...string) (string, error) {
		n := active.Add(1)
		defer active.Add(-1)
		for old := peak.Load(); n > old; old = peak.Load() {
			if peak.CompareAndSwap(old, n) {
				break
			}
		}
		time.Sleep(time.Millisecond)
		if strings.Contains(strings.Join(args, " "), "pkg3@") {
			return "", errors.New("fixture failure")
		}
		return `"2"`, nil
	}
	got, errs := s.findPluginUpdates(context.Background(), pkgs, nil, probe)
	if len(got) != 14 || len(errs) != 1 || peak.Load() > 4 {
		t.Fatal(len(got), errs, peak.Load())
	}
	var wg sync.WaitGroup
	for n := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if e := s.SkipPluginUpdate(fmt.Sprint(n), "2"); e != nil {
				t.Error(e)
			}
		}(n)
	}
	wg.Wait()
	skips, e := s.SkippedPluginUpdates()
	if e != nil || len(skips) != 10 {
		t.Fatal(skips, e)
	}
}
func TestPluginRealLocalGitUpdateDoesNotFetchOrModify(t *testing.T) {
	if _, e := exec.LookPath("git"); e != nil {
		t.Skip(e)
	}
	s := testService(t)
	base := t.TempDir()
	remote := filepath.Join(base, "remote.git")
	origin := filepath.Join(base, "source")
	installed := filepath.Join(base, "installed")
	run := func(dir string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_AUTHOR_NAME=Fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid", "GIT_COMMITTER_NAME=Fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
		out, e := cmd.CombinedOutput()
		if e != nil {
			t.Fatalf("git %v: %v %s", args, e, out)
		}
		return strings.TrimSpace(string(out))
	}
	run(base, "init", "--bare", remote)
	run(base, "init", origin)
	put(t, filepath.Join(origin, "file"), "one")
	run(origin, "add", "file")
	run(origin, "commit", "-m", "one")
	run(origin, "remote", "add", "origin", remote)
	run(origin, "push", "origin", "HEAD:refs/heads/master")
	run(base, "clone", remote, installed)
	before := run(installed, "rev-parse", "HEAD")
	put(t, filepath.Join(origin, "file"), "two")
	run(origin, "add", "file")
	run(origin, "commit", "-m", "two")
	run(origin, "push", "origin", "HEAD:refs/heads/master")
	latest := run(origin, "rev-parse", "HEAD")
	p := Plugin{Source: "git:example.test/a/repo", Name: "repo", Kind: "git", Installed: true, InstallPath: installed}
	update, e := pluginUpdate(context.Background(), p, s.Workspace, nil, s.resourceProbe)
	if e != nil || update == nil || update.Current != before || update.Latest != latest {
		t.Fatal(update, e)
	}
	if after := run(installed, "rev-parse", "HEAD"); after != before {
		t.Fatal("probe modified checkout")
	}
	if remoteRef := run(installed, "rev-parse", "origin/master"); remoteRef != before {
		t.Fatal("probe fetched remote refs")
	}
}
func TestPluginActionsExactArgvAndFailurePersistence(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("temporary POSIX executable fixture")
	}
	s := testService(t)
	bin := filepath.Join(t.TempDir(), "fake-pi")
	log := filepath.Join(t.TempDir(), "argv")
	put(t, bin, "#!/bin/sh\nprintf '%s\\n' \"$@\" > '"+log+"'\nexit 7\n")
	os.Chmod(bin, 0700)
	s.PiBinary = bin
	path := filepath.Join(s.Preferences.PiDir, "settings.json")
	put(t, path, `{"packages":["npm:kept"],"unknown":true}`)
	before, _ := os.ReadFile(path)
	if _, e := s.PluginAction(context.Background(), "remove", "npm:kept", "project"); e == nil {
		t.Fatal("failed CLI reported success")
	}
	args, _ := os.ReadFile(log)
	if !reflect.DeepEqual(strings.Fields(string(args)), []string{"remove", "npm:kept", "-l", "-a"}) {
		t.Fatal(string(args))
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("failed action rewrote settings")
	}
	if _, e := s.PluginAction(context.Background(), "install", "npm:x", "bad"); e == nil {
		t.Fatal("invalid scope")
	}
}

func TestSkillFlatFrontmatterAndExactOverrideSemantics(t *testing.T) {
	name, description, manual := skillFrontmatter("---\rname: 'review'\rdescription: >-\r  Review code\r  carefully\rdisable-model-invocation: TRUE\rmetadata:\r  future: okay\r---\rBody")
	if name != "review" || description != "Review code carefully" || !manual {
		t.Fatal(name, description, manual)
	}
	_, description, _ = skillFrontmatter("---\ndescription:\n  Continued text\n  here\n---\n")
	if description != "Continued text here" {
		t.Fatal(description)
	}
	if skillExactPath("/base/skills/a/SKILL.md", "../base/skills/a/SKILL.md", "/base") {
		t.Fatal("nonexact parent traversal was normalized")
	}
	if !skillExactPath("/base/skills/a/SKILL.md", "./skills/./a/SKILL.md", "/base") {
		t.Fatal("dot components should match source Path equality")
	}
}
