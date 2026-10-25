package mcpui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/charlzyx/orbit/internal/workbench"
)

const secretMask = "••••••••"

var variableName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Draft is a typed native editor, not a raw JSON document. Secret-bearing
// existing literals are masked; original is retained privately for round trips.
type Draft struct {
	Name, Description, Transport, Scope, Exposure                                     string
	Command, Args, Env, URL, Headers, Cwd, Timeout                                    string
	OAuthClientID, OAuthClientSecret, OAuthCallbackPort, OAuthCallbackURL, OAuthScope string
	Enabled, OAuthOpen                                                                bool
	Error                                                                             string
	original                                                                          *workbench.MCPServer
}

type secretUpdate struct{ name, value string }
type preparedDraft struct {
	name, scope, description string
	original                 *workbench.MCPServer
	definition               map[string]any
	secrets                  []secretUpdate
}

func newDraft(server *workbench.MCPServer) *Draft {
	d := &Draft{Transport: "stdio", Scope: "global", Exposure: "codemode", Enabled: true}
	if server == nil {
		return d
	}
	copy := *server
	copy.Definition = cloneMap(server.Definition)
	d.original = &copy
	d.Name, d.Description, d.Scope = server.Name, server.Description, server.Scope
	m := copy.Definition
	d.Enabled = enabled(m)
	d.Exposure = exposure(m)
	d.Command, d.URL, d.Cwd = str(m["command"]), str(m["url"]), str(m["cwd"])
	if d.URL != "" {
		d.Transport = "http"
	}
	if v, ok := m["timeout"]; ok {
		d.Timeout = fmt.Sprint(v)
	}
	var args []string
	for _, a := range array(m["args"]) {
		args = append(args, str(a))
	}
	d.Args = strings.Join(args, "\n")
	d.Env, d.Headers = mapLines(m["env"], "="), mapLines(m["headers"], ": ")
	if o, ok := m["oauth"].(map[string]any); ok {
		d.OAuthOpen = len(o) > 0
		d.OAuthClientID, d.OAuthClientSecret = str(o["clientId"]), displaySecret(str(o["clientSecret"]))
		d.OAuthCallbackURL, d.OAuthScope = str(o["callbackUrl"]), str(o["scope"])
		if v, ok := o["callbackPort"]; ok {
			d.OAuthCallbackPort = fmt.Sprint(v)
		}
	}
	return d
}

func (d *Draft) prepare(workspace string) (preparedDraft, error) {
	out := preparedDraft{name: strings.TrimSpace(d.Name), scope: d.Scope, description: strings.TrimSpace(d.Description), original: d.original, definition: map[string]any{}}
	if d.original != nil {
		out.name = d.original.Name
		out.definition = cloneMap(d.original.Definition)
	}
	if out.scope != "global" && out.scope != "project" {
		return out, errors.New("choose global or project scope")
	}
	if out.scope == "project" && workspace == "" {
		return out, errors.New("select a workspace before saving project configuration")
	}
	for _, k := range []string{"type", "command", "args", "env", "url", "headers", "cwd", "oauth", "timeout"} {
		delete(out.definition, k)
	}
	m := out.definition
	m["enabled"], m["exposure"] = d.Enabled, d.Exposure
	if strings.TrimSpace(d.Timeout) != "" {
		n, e := strconv.ParseUint(strings.TrimSpace(d.Timeout), 10, 53)
		if e != nil || n == 0 {
			return out, errors.New("timeout must be a positive integer")
		}
		m["timeout"] = float64(n)
	}
	resolve := func(field, value, original string) (string, error) {
		if value == secretMask {
			if original == "" {
				return "", fmt.Errorf("%s: a mask is not a secret; enter a value or reference", field)
			}
			value = original
		}
		if value == "" || strings.HasPrefix(value, "!") || strings.Contains(value, "${") {
			return value, nil
		}
		name := generatedSecretName(out.name, field, out.scope, workspace)
		out.secrets = append(out.secrets, secretUpdate{name, value})
		return "${" + name + "}", nil
	}
	resolveMap := func(input, delimiter, key string) (map[string]any, error) {
		parsed, e := parsePairs(input, delimiter, key == "env")
		if e != nil {
			return nil, e
		}
		result := map[string]any{}
		var old map[string]any
		if d.original != nil {
			old, _ = d.original.Definition[key].(map[string]any)
		}
		for _, name := range sortedKeys(parsed) {
			value, e := resolve(name, str(parsed[name]), str(old[name]))
			if e != nil {
				return nil, e
			}
			result[name] = value
		}
		return result, nil
	}
	switch d.Transport {
	case "stdio":
		m["type"], m["command"] = "stdio", strings.TrimSpace(d.Command)
		args := []any{}
		for _, arg := range strings.Split(d.Args, "\n") {
			if arg = strings.TrimSpace(arg); arg != "" {
				args = append(args, arg)
			}
		}
		m["args"] = args
		env, e := resolveMap(d.Env, "=", "env")
		if e != nil {
			return out, e
		}
		if len(env) > 0 {
			m["env"] = env
		}
		if strings.TrimSpace(d.Cwd) != "" {
			m["cwd"] = strings.TrimSpace(d.Cwd)
		}
	case "http":
		m["type"], m["url"] = "http", strings.TrimSpace(d.URL)
		headers, e := resolveMap(d.Headers, ":", "headers")
		if e != nil {
			return out, e
		}
		if len(headers) > 0 {
			m["headers"] = headers
		}
		oauth := map[string]any{}
		for key, value := range map[string]string{"clientId": d.OAuthClientID, "callbackUrl": d.OAuthCallbackURL, "scope": d.OAuthScope} {
			if value = strings.TrimSpace(value); value != "" {
				oauth[key] = value
			}
		}
		if d.OAuthCallbackPort != "" {
			n, e := strconv.Atoi(strings.TrimSpace(d.OAuthCallbackPort))
			if e != nil || n < 1 || n > 65535 {
				return out, errors.New("OAuth callback port must be 1–65535")
			}
			oauth["callbackPort"] = float64(n)
		}
		if raw := str(oauth["callbackUrl"]); raw != "" {
			u, e := url.Parse(raw)
			if e != nil || u.Scheme != "http" || (u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1" && u.Hostname() != "::1") {
				return out, errors.New("OAuth callback URL must use loopback HTTP")
			}
		}
		original := ""
		if d.original != nil {
			old, _ := d.original.Definition["oauth"].(map[string]any)
			original = str(old["clientSecret"])
		}
		value, e := resolve("OAuthClientSecret", strings.TrimSpace(d.OAuthClientSecret), original)
		if e != nil {
			return out, e
		}
		if value != "" {
			oauth["clientSecret"] = value
		}
		if len(oauth) > 0 {
			m["oauth"] = oauth
		}
	default:
		return out, errors.New("choose stdio or streamable HTTP transport")
	}
	return out, workbench.ValidateMCP(out.name, m)
}

func parsePairs(text, delimiter string, env bool) (map[string]any, error) {
	out := map[string]any{}
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, delimiter)
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf("line %d must be %s", i+1, map[bool]string{true: "NAME=value", false: "Header: value"}[env])
		}
		if env && !variableName.MatchString(key) {
			return nil, fmt.Errorf("line %d has an invalid environment variable name", i+1)
		}
		if !env && strings.ContainsAny(key, " \t\r\n:") {
			return nil, fmt.Errorf("line %d has an invalid header name", i+1)
		}
		out[key] = strings.TrimSpace(value)
	}
	return out, nil
}
func displaySecret(value string) string {
	if value == "" || strings.HasPrefix(value, "!") || strings.Contains(value, "${") {
		return value
	}
	return secretMask
}
func mapLines(raw any, separator string) string {
	m, _ := raw.(map[string]any)
	var lines []string
	for _, key := range sortedKeys(m) {
		lines = append(lines, key+separator+displaySecret(str(m[key])))
	}
	return strings.Join(lines, "\n")
}
func secretNamespace(scope, workspace string) string {
	return workbench.MCPSecretNamespace(scope, workspace)
}
func generatedSecretName(server, field, scope, workspace string) string {
	return workbench.MCPGeneratedSecretName(server, field, scope, workspace)
}
func cloneMap(m map[string]any) map[string]any {
	b, _ := json.Marshal(m)
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	if out == nil {
		out = map[string]any{}
	}
	return out
}
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func str(v any) string              { s, _ := v.(string); return s }
func array(v any) []any             { s, _ := v.([]any); return s }
func enabled(m map[string]any) bool { on, ok := m["enabled"].(bool); return !ok || on }
func exposure(m map[string]any) string {
	if v := str(m["exposure"]); v != "" {
		return v
	}
	return "codemode"
}
func oauthCandidate(server workbench.MCPServer) bool {
	if str(server.Definition["url"]) == "" {
		return false
	}
	headers, _ := server.Definition["headers"].(map[string]any)
	for key := range headers {
		if strings.EqualFold(key, "Authorization") {
			return false
		}
	}
	return true
}
