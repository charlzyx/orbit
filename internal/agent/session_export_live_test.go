package agent

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLiveSessionExportExactTargetNoInference(t *testing.T) {
	binary := os.Getenv("ORBIT_PI_INTEGRATION_BIN")
	if binary == "" {
		t.Skip("set ORBIT_PI_INTEGRATION_BIN for pinned credential-free export verification")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	home, workspace := t.TempDir(), t.TempDir()
	agentDir := filepath.Join(home, ".pi", "agent")
	if err = os.MkdirAll(agentDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(agentDir, "settings.json"), []byte(`{"enableInstallTelemetry":false,"enableAnalytics":false}`), 0600); err != nil {
		t.Fatal(err)
	}
	env := []string{"PATH=" + filepath.Dir(node) + string(os.PathListSeparator) + os.Getenv("PATH"), "HOME=" + home, "PI_CODING_AGENT_DIR=" + agentDir, "PI_OFFLINE=1", "PI_SKIP_VERSION_CHECK=1"}
	version := exec.Command(binary, "--version")
	version.Env = env
	raw, err := version.Output()
	if err != nil || string(raw) != "1.4.0\n" {
		t.Fatalf("expected Pi1.4.0: %q %v", raw, err)
	}
	const id = "01900000-0000-7000-8000-000000000001"
	const marker = "Orbit export fixture 9bc57c 中文"
	path := filepath.Join(workspace, "saved session.jsonl")
	header, _ := json.Marshal(map[string]any{"type": "session", "version": 3, "id": id, "cwd": workspace, "timestamp": "2026-01-01T00:00:00Z"})
	message, _ := json.Marshal(map[string]any{"type": "message", "id": "entry123", "parentId": nil, "timestamp": "2026-01-01T00:00:00Z", "message": map[string]any{"role": "user", "content": []any{map[string]any{"type": "text", "text": marker}}, "timestamp": 1767225600000}})
	content := append(append(append(header, '\n'), message...), '\n')
	if err = os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	m := NewManager(Options{PiBinary: binary, Home: home, Environment: env, DisableExtensions: true, DisablePatches: true, DisableCheckpoints: true, Offline: true, CommandTimeout: 15 * time.Second, PiArgs: []string{"--no-extensions", "--no-skills", "--no-prompt-templates", "--no-context-files", "--no-tools", "--no-approve"}})
	defer m.Close()
	if err = m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID == id })
	target, err := m.CaptureSessionExportTarget()
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "真实 export.html")
	if err = m.ExportSessionFor(target, output); err != nil {
		t.Fatal(err)
	}
	m.mu.Lock()
	entry := m.sessions[m.active]
	m.mu.Unlock()
	waitExportLeaseGone(t, m, entry)
	html, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("export output missing: %v; snapshot: %+v", err, m.Snapshot())
	}
	// Pinned Pi embeds base64 JSON in its session-data script, not visible text.
	_, embedded, ok := strings.Cut(string(html), `<script id="session-data" type="application/json">`)
	if !ok {
		t.Fatal("missing session-data payload")
	}
	encoded, _, ok := strings.Cut(embedded, "</script>")
	if !ok {
		t.Fatal("unterminated session-data payload")
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		t.Fatal(err)
	}
	var exported struct {
		Header struct {
			ID string `json:"id"`
		}
		Entries []json.RawMessage
	}
	if err = json.Unmarshal(decoded, &exported); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(string(html)), "<!doctype html") || exported.Header.ID != id || !strings.Contains(string(decoded), marker) {
		t.Fatalf("not the selected saved session HTML (%d bytes; ID %s)", len(html), exported.Header.ID)
	}
	if s := m.Snapshot(); s.Running || s.Awaiting {
		t.Fatal("metadata export unexpectedly started inference")
	}
	beforePID := target.client.PID()
	if err = m.NewSession(workspace); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" && s.ActiveSession.ID != id })
	waitExportLeaseGone(t, m, entry)
	fresh, err := m.CaptureSessionExportTarget()
	if err != nil {
		t.Fatal(err)
	}
	if fresh.client.PID() != beforePID {
		t.Fatal("idle new_session did not test same-child reuse")
	}
	if err = m.ExportSessionFor(target, output); err == nil {
		t.Fatal("old captured target exported newer task")
	}
	after, err := os.ReadFile(output)
	if err != nil || string(after) != string(html) {
		t.Fatal("rejected stale target changed prior HTML")
	}
}
