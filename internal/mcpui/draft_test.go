package mcpui

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/charlzyx/orbit/internal/workbench"
)

func TestTypedDraftStdioAndSecretReferences(t *testing.T) {
	d := newDraft(nil)
	d.Name = "github"
	d.Command = "npx"
	d.Args = "-y\n@modelcontextprotocol/server-github\n"
	d.Env = "TOKEN=not-a-real-token\nEXISTING=${MY_TOKEN}\n# comment\n"
	prepared, e := d.prepare("/work/project")
	if e != nil {
		t.Fatal(e)
	}
	if len(prepared.secrets) != 1 || prepared.secrets[0].name != "MCP_GLOBAL_GITHUB_TOKEN" {
		t.Fatalf("unexpected generated names: %#v", prepared.secrets)
	}
	encoded, _ := json.Marshal(prepared.definition)
	if strings.Contains(string(encoded), "not-a-real-token") || !strings.Contains(string(encoded), "${MCP_GLOBAL_GITHUB_TOKEN}") {
		t.Fatal("literal secret entered Pi configuration")
	}
	if prepared.definition["env"].(map[string]any)["EXISTING"] != "${MY_TOKEN}" {
		t.Fatal("existing reference changed")
	}
	if len(prepared.definition["args"].([]any)) != 2 {
		t.Fatal("one argument per line was not preserved")
	}
}
func TestDraftMasksAndPreservesUnknownAndToolFields(t *testing.T) {
	server := workbench.MCPServer{Name: "one", Scope: "global", Definition: map[string]any{"command": "node", "env": map[string]any{"TOKEN": "old-literal"}, "toolExposure": map[string]any{"read": "direct"}, "custom": map[string]any{"future": true}, "timeout": float64(900)}}
	d := newDraft(&server)
	if strings.Contains(d.Env, "old-literal") || !strings.Contains(d.Env, secretMask) {
		t.Fatal("literal secret shown in editor")
	}
	d.Name = "cannot-rename"
	p, e := d.prepare("")
	if e != nil {
		t.Fatal(e)
	}
	if p.name != "one" || p.definition["toolExposure"] == nil || p.definition["custom"] == nil || p.definition["timeout"] != float64(900) {
		t.Fatalf("fields not preserved: %+v", p.definition)
	}
	if len(p.secrets) != 1 || p.secrets[0].value != "old-literal" {
		t.Fatal("untouched mask did not preserve secret")
	}
}
func TestHTTPDraftOAuthValidationAndTransportSwitch(t *testing.T) {
	d := newDraft(nil)
	d.Name = "remote"
	d.Transport = "http"
	d.URL = "https://example.invalid/mcp"
	d.Headers = "Authorization: Bearer ${TOKEN}\nX-Value: abc:def"
	d.OAuthClientID = "id"
	d.OAuthClientSecret = "client-secret"
	d.OAuthCallbackPort = "7777"
	d.OAuthCallbackURL = "http://[::1]:7777/callback"
	p, e := d.prepare("/project")
	if e != nil {
		t.Fatal(e)
	}
	if p.definition["command"] != nil || p.definition["oauth"].(map[string]any)["callbackPort"] != float64(7777) || len(p.secrets) != 2 {
		t.Fatal("incorrect HTTP form")
	}
	d.OAuthCallbackURL = "https://remote.example/callback"
	if _, e = d.prepare("/project"); e == nil {
		t.Fatal("non-loopback OAuth redirect accepted")
	}
	d.OAuthCallbackURL = "http://localhost:7777"
	d.OAuthCallbackPort = "65536"
	if _, e = d.prepare(""); e == nil {
		t.Fatal("invalid OAuth port accepted")
	}
}
func TestGeneratedSecretsScopeNamespaceMatchesUpstream(t *testing.T) {
	if got := generatedSecretName("my-server", "Authorization", "global", ""); got != "MCP_GLOBAL_MY_SERVER_AUTHORIZATION" {
		t.Fatal(got)
	}
	a := generatedSecretName("same", "TOKEN", "project", "/project/a")
	b := generatedSecretName("same", "TOKEN", "project", "/project/b")
	if a == b || a == generatedSecretName("same", "TOKEN", "global", "") {
		t.Fatal("scopes share secret storage")
	}
	if got := secretNamespace("project", ""); got != "MCP_PCBF29CE484222325_" {
		t.Fatal("namespace does not use source-compatible FNV1a", got)
	}
}
func TestTypedDraftRejectsInvalidFields(t *testing.T) {
	for _, test := range []struct {
		name string
		edit func(*Draft)
	}{
		{"name", func(d *Draft) { d.Name = "bad name" }},
		{"project without workspace", func(d *Draft) { d.Scope = "project" }},
		{"environment", func(d *Draft) { d.Env = "missing delimiter" }},
		{"variable", func(d *Draft) { d.Env = "NOT-VALID=value" }},
		{"timeout", func(d *Draft) { d.Timeout = "1.5" }},
		{"mask without original", func(d *Draft) { d.Env = "TOKEN=" + secretMask }},
	} {
		t.Run(test.name, func(t *testing.T) {
			d := newDraft(nil)
			d.Name = "valid"
			d.Command = "node"
			test.edit(d)
			if _, e := d.prepare(""); e == nil {
				t.Fatal("invalid draft accepted")
			}
		})
	}
}
func TestProbeParsesTruthAndUnknownStates(t *testing.T) {
	p, e := parseProbe([]byte(`{"servers":[{"name":"one","state":"connected","tools":["read",{"name":"write","description":"Writes a file","inputSchema":{"type":"object"}}],"toolExposure":{"read":"direct"},"resources":2,"resourceTemplates":1},{"name":"unknown","state":"futuristic"}],"errors":["configuration failed"],"note":"Project is untrusted"}`))
	if e != nil {
		t.Fatal(e)
	}
	r := p.Servers["one"]
	if r.State != "connected" || len(r.Tools) != 2 || r.Tools[0].Exposure != "direct" || r.Tools[1].Description == "" || r.Resources != 2 || r.ResourceTemplates != 1 {
		t.Fatalf("probe data lost: %+v", r)
	}
	if p.Servers["unknown"].State != "unknown" || len(p.Errors) != 1 || p.Note == "" {
		t.Fatal("probe truth fabricated or dropped")
	}
	if _, e := parseProbe([]byte(`{"state":"connected"}`)); e == nil {
		t.Fatal("missing servers accepted")
	}
}
