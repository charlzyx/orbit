package native

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo"
)

func TestNotificationPreferencesMigration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notifications.json")
	p, err := LoadNotificationPrefs(path)
	if err != nil || p != DefaultNotificationPrefs() {
		t.Fatal(p, err)
	}
	if err = os.WriteFile(path, []byte(`{"desktop":false,"sound":"invalid","toasts":null}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err = LoadNotificationPrefs(path)
	if err != nil || p.Desktop || !p.Sound || !p.Toasts {
		t.Fatal(p, err)
	}
	p.Toasts = false
	if err = p.Save(path); err != nil {
		t.Fatal(err)
	}
	again, err := LoadNotificationPrefs(path)
	if err != nil || again != p {
		t.Fatal(again, err)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0600 {
		t.Fatalf("prefs mode %o", info.Mode())
	}
}
func TestPreviewUnicodeAndControls(t *testing.T) {
	if got := Preview(" Hello\n\t world\x00! ", 30); got != "Hello world!" {
		t.Fatal(got)
	}
	if got := Preview("世界你好世界", 4); got != "世界你好…" {
		t.Fatal(got)
	}
	if got := Preview("one two three four", 12); got != "one two…" {
		t.Fatal(got)
	}
	if got := Preview("hello", 0); got != "…" {
		t.Fatal(got)
	}
}
func TestNotificationChannelPolicyAndClickRoute(t *testing.T) {
	n := NewNotifier(DefaultNotificationPrefs())
	var banner mygo.NotificationOptions
	sounds, toasts := 0, 0
	n.ShowBanner = func(o mygo.NotificationOptions) error { banner = o; return nil }
	n.PlaySound = func() { sounds++ }
	n.OnToast = func(k ToastKind, title, body string) { toasts++ }
	event := Notification{Session: "/tmp/session.jsonl", Body: "done", Kind: ToastSuccess}
	n.SetForeground(true)
	r := n.Deliver(event)
	if !r.Toast || r.Banner || r.Sound || sounds != 0 || toasts != 1 {
		t.Fatal(r)
	}
	n.SetForeground(false)
	r = n.Deliver(event)
	if r.Toast || !r.Banner || !r.Sound || sounds != 1 || banner.Title != "Orbit Pi" || !banner.Silent {
		t.Fatal(r, banner)
	}
	if session, ok := SessionFromNotification(banner.ID); !ok || session != event.Session {
		t.Fatal(session, ok)
	}
	if _, ok := SessionFromNotification("other:session"); ok {
		t.Fatal("foreign notification routed")
	}
	n.ShowBanner = func(mygo.NotificationOptions) error { return errors.New("permission denied") }
	if got := n.Deliver(event); got.Error == nil || got.Banner {
		t.Fatal(got)
	}
}
func TestToastCapDedupeExpiryAndReducedMotion(t *testing.T) {
	s := &ToastStack{}
	now := time.Unix(1, 0)
	id := s.PushAt(ToastSuccess, "Copied", "", now)
	if next := s.PushAt(ToastSuccess, "Copied", "", now.Add(time.Second)); next != id || len(s.Items()) != 1 || s.Items()[0].Revision != 1 {
		t.Fatal(s.Items())
	}
	for _, title := range []string{"a", "b", "c", "d"} {
		s.PushAt(ToastInfo, title, "", now)
	}
	if len(s.Items()) != 4 || s.Items()[0].Title != "a" {
		t.Fatal(s.Items())
	}
	if s.Expire(now.Add(4*time.Second), false) {
		t.Fatal("removed before fade")
	}
	if !s.Expire(now.Add(4*time.Second), true) || len(s.Items()) != 0 {
		t.Fatal("reduce-motion expiry")
	}
	v := Toast{Kind: ToastSuccess, Created: now}
	if opacity := v.Opacity(now.Add(3*time.Second+ToastFade/2), false); opacity < 0.49 || opacity > 0.51 {
		t.Fatal(opacity)
	}
}
func TestOpenInKeepsUnavailableWorkspaceChoice(t *testing.T) {
	path := filepath.Join(t.TempDir(), "open-in.json")
	prefs := OpenInPrefs{OpenInApp: "zed"}
	prefs.Remember("/work", "cursor")
	apps := []ExternalApp{{ID: "vscode"}, {ID: "zed"}}
	if got := prefs.Preferred("/work", apps); got == nil || got.ID != "zed" {
		t.Fatal(got)
	}
	if prefs.Workspaces["/work"] != "cursor" {
		t.Fatal("fallback overwrote choice")
	}
	if err := prefs.Save(path); err != nil {
		t.Fatal(err)
	}
	p, err := LoadOpenInPrefs(path)
	if err != nil {
		t.Fatal(err)
	}
	apps = append(apps, ExternalApp{ID: "cursor"})
	if got := p.Preferred("/work", apps); got == nil || got.ID != "cursor" {
		t.Fatal(got)
	}
	raw, _ := os.ReadFile(path)
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields["open_in_app"] == nil {
		t.Fatal(string(raw))
	}
}
func TestTermyDirectoryEscaping(t *testing.T) {
	got := TermyURL("/tmp/hello & 世界/#a")
	if !strings.Contains(got, "termy://new?dir=") || strings.Contains(got, " #") {
		t.Fatal(got)
	}
}
func TestEnvironmentExpansion(t *testing.T) {
	t.Setenv("ORBIT_TEST_APP_ROOT", "C:\\Tools")
	got, ok := expandWindowsEnvironment("%ORBIT_TEST_APP_ROOT%\\app.exe")
	if !ok || got != "C:\\Tools\\app.exe" {
		t.Fatal(got, ok)
	}
	if _, ok = expandWindowsEnvironment("%NOT_DEFINED_ORBIT_TEST%\\x"); ok {
		t.Fatal("missing environment accepted")
	}
}

func TestRuntimeHostFactsAreReadOnly(t *testing.T) {
	root := t.TempDir()
	host := HostInfo{Label: "Fixture OS", Arch: "fixture64", Unsupported: "Needs newer OS"}
	facts := RuntimeHostFacts(host, filepath.Join(root, "agent"), filepath.Join(root, "orbit"))
	if len(facts) != 3 || !facts[0].Alert || !strings.Contains(facts[0].Value, "Fixture OS") {
		t.Fatal(facts)
	}
	for _, f := range facts[1:] {
		if f.Alert || f.Note != "not created yet" {
			t.Fatal(f)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "agent")); !os.IsNotExist(err) {
		t.Fatal("host facts wrote to disk")
	}
	if got := parseOSRelease("NAME=Ubuntu\nPRETTY_NAME=\"Ubuntu 24.04.1 LTS\"\n"); got != "Ubuntu 24.04.1 LTS" {
		t.Fatal(got)
	}
}
