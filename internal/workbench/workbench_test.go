package workbench

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/charlzyx/orbit/internal/preferences"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func testService(t *testing.T) *Service {
	t.Helper()
	home := t.TempDir()
	workspace := filepath.Join(home, "project")
	os.MkdirAll(workspace, 0700)
	return New(preferences.New(home), workspace)
}
func put(t *testing.T, path, content string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0700)
	if e := os.WriteFile(path, []byte(content), 0600); e != nil {
		t.Fatal(e)
	}
}
func TestProviderSurgicalUpdate(t *testing.T) {
	s := testService(t)
	p := filepath.Join(s.Preferences.PiDir, "models.json")
	put(t, p, `{"customTop":true,"providers":{"local":{"apiKey":"SECRET","contextPolicy":"keep","models":[{"id":"m","contextWindow":1234,"cost":{"input":1}}]},"other":{"baseUrl":"http://localhost:9999"}}}`)
	if e := s.SaveProvider(ProviderEdit{ID: "local", BaseURL: "http://localhost:11434", API: "openai-completions", ModelIDs: []string{"m", "new"}}); e != nil {
		t.Fatal(e)
	}
	m, _ := readObject(p)
	entry := m["providers"].(map[string]any)["local"].(map[string]any)
	if m["customTop"] != true || entry["apiKey"] != "SECRET" || entry["contextPolicy"] != "keep" || entry["models"].([]any)[0].(map[string]any)["contextWindow"] != float64(1234) {
		t.Fatal("unrelated configuration lost")
	}
	ps, e := s.Providers()
	if e != nil || len(ps) != 2 || !ps[0].HasAPIKey {
		t.Fatalf("%+v %v", ps, e)
	}
	b, _ := json.Marshal(ps)
	if bytes.Contains(b, []byte("SECRET")) {
		t.Fatal("credentials entered UI snapshot")
	}
	s.RemoveProvider("local")
	m, _ = readObject(p)
	if m["customTop"] != true || len(m["providers"].(map[string]any)) != 1 {
		t.Fatal("remove touched other keys")
	}
}
func TestProviderBadJSONCannotBeOverwritten(t *testing.T) {
	s := testService(t)
	p := filepath.Join(s.Preferences.PiDir, "models.json")
	put(t, p, "{broken")
	if e := s.SaveProvider(ProviderEdit{ID: "x", BaseURL: "http://localhost"}); e == nil {
		t.Fatal("overwrote malformed config")
	}
	b, _ := os.ReadFile(p)
	if string(b) != "{broken" {
		t.Fatal("data changed")
	}
}
func TestCredentialMetadataOnlyAndSecureMode(t *testing.T) {
	s := testService(t)
	if e := s.SaveAPIKey("test", "not-a-real-key"); e != nil {
		t.Fatal(e)
	}
	statuses, e := s.Credentials()
	if e != nil || !statuses["test"].Configured || statuses["test"].Kind != "api_key" {
		t.Fatal(statuses, e)
	}
	info, _ := os.Stat(filepath.Join(s.Preferences.PiDir, "auth.json"))
	if info.Mode().Perm()&0077 != 0 {
		t.Fatal("unsafe credential permissions")
	}
}
func TestMCPProjectMergeAndUnknownPreservation(t *testing.T) {
	s := testService(t)
	g := filepath.Join(s.Preferences.PiDir, "mcp.json")
	p := filepath.Join(s.Workspace, ".pi", "mcp.json")
	put(t, g, `{"top":"keep","mcpServers":{"same":{"command":"global","other":42},"bad":{"type":"sse","url":"https://example.test"}}}`)
	put(t, p, `{"mcpServers":{"same":{"command":"project"}}}`)
	servers, errs := s.MCP()
	if len(servers) != 1 || servers[0].Scope != "project" || len(errs) != 1 {
		t.Fatal(servers, errs)
	}
	if e := s.SaveMCP("global", "same", "renamed", map[string]any{"command": "updated", "enabled": false}); e != nil {
		t.Fatal(e)
	}
	m, _ := readObject(g)
	entries := m["mcpServers"].(map[string]any)
	if m["top"] != "keep" || entries["renamed"].(map[string]any)["other"] != float64(42) || entries["same"] != nil || entries["bad"] == nil {
		t.Fatal("MCP edit was destructive")
	}
}
func TestMCPValidation(t *testing.T) {
	bad := []map[string]any{{"type": "sse", "url": "https://example.test"}, {"command": ""}, {"command": "x", "args": []any{1}}, {"url": "file:///private"}, {"url": "https://example.test", "type": "stdio"}, {"command": "x", "timeout": float64(0)}, {"command": "x", "enabled": "true"}, {"command": "x", "env": map[string]any{"KEY": 1}}, {"url": "https://example.test", "oauth": map[string]any{"callbackUrl": "https://elsewhere.test"}}}
	for _, def := range bad {
		if e := ValidateMCP("test", def); e == nil {
			t.Errorf("accepted %+v", def)
		}
	}
	if e := ValidateMCP("ok", map[string]any{"command": "node", "args": []any{"server.js"}, "exposure": "direct"}); e != nil {
		t.Fatal(e)
	}
}
func TestSecretsRedactionAndCorruptStore(t *testing.T) {
	s := testService(t)
	if e := s.SaveSecret("TOKEN", "secret-value"); e != nil {
		t.Fatal(e)
	}
	env, e := s.SecretEnv()
	if e != nil || !reflect.DeepEqual(env, []string{"TOKEN=secret-value"}) {
		t.Fatal(env, e)
	}
	got := s.Redact("secret-value Bearer unknown-token")
	if strings.Contains(got, "secret-value") || strings.Contains(got, "unknown-token") {
		t.Fatal(got)
	}
	put(t, filepath.Join(s.Preferences.RootDir, "mcp-secrets.json"), "{broken")
	if s.SaveSecret("OTHER", "x") == nil {
		t.Fatal("overwrote corrupt secrets")
	}
	if _, e := s.SecretEnv(); e == nil {
		t.Fatal("silently dropped unreadable secrets")
	}
	if got := SecretReferences("${A} ${B} ${A}"); !reflect.DeepEqual(got, []string{"A", "B"}) {
		t.Fatal(got)
	}
}
func TestSkillPrecedenceFrontmatterAndOverride(t *testing.T) {
	s := testService(t)
	global := filepath.Join(s.Preferences.PiDir, "skills", "same", "SKILL.md")
	project := filepath.Join(s.Workspace, ".pi", "skills", "same", "SKILL.md")
	put(t, global, "---\nname: shared\ndescription: global\n---\nGlobal.")
	put(t, project, "---\nname: shared\ndescription: |\n  Multiple lines\n  for this skill.\ndisable-model-invocation: true\n---\nProject.")
	skills, errs := s.Skills()
	if len(errs) > 0 || len(skills) != 1 || skills[0].Scope != "project" || !skills[0].ManualOnly || !strings.Contains(skills[0].Description, "Multiple lines") {
		t.Fatal(skills, errs)
	}
	if e := s.SetSkillEnabled(skills[0], false); e != nil {
		t.Fatal(e)
	}
	skills, _ = s.Skills()
	if skills[0].Enabled {
		t.Fatal("disable not persisted")
	}
	s.SetSkillEnabled(skills[0], true)
	skills, _ = s.Skills()
	if !skills[0].Enabled {
		t.Fatal("enable not restored")
	}
	if e := s.DeleteSkill(filepath.Join(s.Preferences.Home, "outside", "SKILL.md")); e == nil {
		t.Fatal("deleted undiscovered path")
	}
}
func TestPluginProjectPrecedenceAndResolution(t *testing.T) {
	s := testService(t)
	put(t, filepath.Join(s.Preferences.PiDir, "settings.json"), `{"packages":["npm:@scope/pkg@1.0","github:egoist/repo","./local"]}`)
	put(t, filepath.Join(s.Workspace, ".pi", "settings.json"), `{"packages":[{"source":"npm:@scope/pkg@1.0"}]}`)
	list, errs := s.Plugins()
	if len(errs) != 0 || len(list) != 3 {
		t.Fatal(list, errs)
	}
	for _, p := range list {
		if p.Kind == "npm" && (p.Scope != "project" || p.Name != "@scope/pkg" || !strings.Contains(p.InstallPath, filepath.Join(".pi", "npm", "node_modules", "@scope", "pkg"))) {
			t.Fatal(p)
		}
	}
	if _, e := s.PluginAction(context.Background(), "install", "--malicious", "global"); e == nil {
		t.Fatal("option injection")
	}
}
func usageFixture() string {
	return `{"type":"session","id":"s1","cwd":"/work","timestamp":"2026-10-25T01:00:00Z"}
{"type":"message","timestamp":"2026-10-25T01:00:00Z","message":{"role":"user","content":"hi"}}
{"type":"message","timestamp":"2026-10-25T01:00:02Z","message":{"role":"assistant","provider":"test","model":"m","stopReason":"toolUse","usage":{"input":10,"output":20,"cacheRead":30,"cacheWrite":40,"reasoning":5,"cost":{"total":0.01}},"content":[{"type":"toolCall","id":"t1","name":"bash"}]}}
{"type":"message","timestamp":"2026-10-25T01:00:03Z","message":{"role":"toolResult","toolCallId":"t1","toolName":"bash","isError":false}}
{"type":"message","timestamp":"2026-10-25T01:00:05Z","message":{"role":"assistant","provider":"test","model":"m","stopReason":"stop","usage":{"input":2,"output":3,"totalTokens":5,"cost":{"total":0.02}}}}
{"type":"message","timestamp":"2026-10-25T01:00:06Z","message":{"role":"assistant","usage":{"input":0,"output":0,"cost":{"total":0}}}}
{partial`
}
func TestUsageRealFileAccountingCachingAndCSV(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "group", "s.jsonl")
	put(t, file, usageFixture())
	scanner := NewUsageScanner(root)
	idx, e := scanner.Scan(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	s := idx.Summary(UsageFilter{})
	if s.Requests != 2 || s.Turns != 1 || s.ToolRuns != 1 || s.BashRuns != 1 || s.Tokens.Total != 105 || s.Tokens.Reasoning != 5 || s.Cost != 0.03 || s.Duration != 4*time.Second {
		t.Fatalf("%+v", s)
	}
	if idx.Tools[0].Duration != time.Second {
		t.Fatal(idx.Tools)
	}
	idx, e = scanner.Scan(context.Background())
	if e != nil || idx.Parsed != 0 {
		t.Fatal("unchanged file rescanned", e)
	}
	var b bytes.Buffer
	if e = idx.ExportCSV(&b, UsageFilter{Provider: "test"}); e != nil || strings.Count(b.String(), "\n") != 3 {
		t.Fatal(b.String(), e)
	}
	series := idx.Series(UsageFilter{}, "day", time.UTC)
	if len(series) != 1 || series[0].Tokens != 105 {
		t.Fatal(series)
	}
	os.Remove(file)
	idx, _ = scanner.Scan(context.Background())
	if len(idx.Records) != 0 {
		t.Fatal("deleted file stayed in cache")
	}
}
func TestUsageTimeAndInjectionBoundaries(t *testing.T) {
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	start, end := UsageRange("last7", now)
	if start.Day() != 19 || !end.Equal(now) {
		t.Fatal(start, end)
	}
	start, end = UsageRange("previous-month", now)
	if start.Month() != time.September || end.Month() != time.October {
		t.Fatal(start, end)
	}
	if csvSafe("=formula") != "'=formula" || csvSafe("") != "" {
		t.Fatal("CSV injection guard")
	}
	tokens := Tokens{Input: 10, CacheRead: 30, CacheWrite: 10}
	if tokens.CacheRate() != 60 {
		t.Fatal(tokens.CacheRate())
	}
}
func TestRetentionRequiresOptInAndSkipsLiveSessions(t *testing.T) {
	s := testService(t)
	path := filepath.Join(s.Preferences.PiDir, "sessions", "w", "s.jsonl")
	put(t, path, strings.ReplaceAll(usageFixture(), "2026-10-25", "2026-01-01"))
	now := time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
	out, e := s.ApplyRetention(now, nil)
	if e != nil || out.Archived != 0 {
		t.Fatal(out, e)
	}
	s.Preferences.SaveRetention(preferences.Retention{Enabled: true, Mode: "archive", Days: 30})
	out, e = s.ApplyRetention(now, []string{path})
	if e != nil || len(out.SkippedLive) != 1 || out.Archived != 0 {
		t.Fatal(out, e)
	}
	out, e = s.ApplyRetention(now, nil)
	if e != nil || out.Archived != 1 || len(s.Preferences.SessionSet(true)) != 1 {
		t.Fatal(out, e)
	}
	if _, e = os.Stat(path); e != nil {
		t.Fatal("archive removed Pi session")
	}
	s.Preferences.SetSession(false, path, true)
	s.Preferences.SaveRetention(preferences.Retention{Enabled: true, Mode: "delete", Days: 30})
	out, e = s.ApplyRetention(now, nil)
	if e != nil || out.Deleted != 1 || len(s.Preferences.SessionSet(false)) != 0 || len(s.Preferences.SessionSet(true)) != 0 {
		t.Fatal(out, e)
	}
}
