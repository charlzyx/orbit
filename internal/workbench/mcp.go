package workbench

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io/fs"
	"math"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

type MCPServer struct {
	Name, Scope, Source, Description string
	Definition                       map[string]any
}

var MCPExposures = []string{"codemode", "codemode-deferred", "deferred", "direct", "hidden"}
var mcpName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
var secretReference = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)
var bearer = regexp.MustCompile(`(?i)Bearer\s+[^\s"']+`)

func (s *Service) mcpPath(scope string) (string, error) {
	p, e := s.settingsPath(scope)
	if e != nil {
		return "", e
	}
	return filepath.Join(filepath.Dir(p), "mcp.json"), nil
}
func ValidateMCP(name string, def map[string]any) error {
	if !mcpName.MatchString(name) {
		return errors.New("server name must contain only letters, digits, hyphen or underscore")
	}
	if def == nil {
		return errors.New("server definition must be an object")
	}
	kind := stringValue(def["type"])
	if v, ok := def["type"]; ok {
		if _, ok := v.(string); !ok {
			return errors.New("type must be a string")
		}
	}
	if kind == "sse" {
		return errors.New("legacy SSE is unsupported")
	}
	if v, ok := def["enabled"]; ok {
		if _, ok := v.(bool); !ok {
			return errors.New("enabled must be boolean")
		}
	}
	if v, ok := def["exposure"]; ok {
		if !slices.Contains(MCPExposures, stringValue(v)) {
			return errors.New("invalid exposure mode")
		}
	}
	if v, ok := def["timeout"]; ok {
		f, ok := mcpNumber(v)
		if !ok || f <= 0 || math.IsInf(f, 0) || math.IsNaN(f) || math.Trunc(f) != f {
			return errors.New("timeout must be a positive integer")
		}
	}
	if v, ok := def["toolExposure"]; ok {
		m, ok := v.(map[string]any)
		if !ok {
			return errors.New("toolExposure must be an object")
		}
		for _, v := range m {
			if !slices.Contains(MCPExposures, stringValue(v)) {
				return errors.New("invalid tool exposure mode")
			}
		}
	}
	if raw, ok := def["url"]; ok {
		u, e := url.Parse(stringValue(raw))
		if e != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return errors.New("server URL must be absolute http or https")
		}
		if kind != "" && kind != "http" && kind != "streamable-http" {
			return errors.New("transport type does not match URL")
		}
		if rawOAuth, exists := def["oauth"]; exists {
			oauth, ok := rawOAuth.(map[string]any)
			if !ok {
				return errors.New("oauth must be an object")
			}
			for _, key := range []string{"clientId", "clientSecret", "callbackUrl", "scope"} {
				if value, exists := oauth[key]; exists {
					if _, ok := value.(string); !ok {
						return fmt.Errorf("oauth.%s must be a string", key)
					}
				}
			}
			if rawPort, exists := oauth["callbackPort"]; exists {
				port, ok := mcpNumber(rawPort)
				if !ok || port < 1 || port > 65535 || math.Trunc(port) != port {
					return errors.New("OAuth callback port must be 1–65535")
				}
			}
			if callback := stringValue(oauth["callbackUrl"]); callback != "" {
				u, e := url.Parse(callback)
				if e != nil || u.Scheme != "http" || !slices.Contains([]string{"localhost", "127.0.0.1", "::1"}, u.Hostname()) {
					return errors.New("OAuth callback must use loopback HTTP")
				}
			}
		}
	} else {
		command := stringValue(def["command"])
		if strings.TrimSpace(command) == "" || strings.ContainsRune(command, 0) {
			return errors.New("a valid command is required")
		}
		if kind != "" && kind != "stdio" {
			return errors.New("transport type does not match command")
		}
		if v, ok := def["args"]; ok {
			a, ok := v.([]any)
			if !ok {
				return errors.New("args must be an array")
			}
			for _, a := range a {
				if _, ok := a.(string); !ok {
					return errors.New("args must contain strings")
				}
			}
		}
		if v, ok := def["cwd"]; ok {
			str, ok := v.(string)
			if !ok || strings.ContainsRune(str, 0) {
				return errors.New("cwd must be a valid string")
			}
		}
	}
	for _, key := range []string{"env", "headers"} {
		if v, ok := def[key]; ok {
			m, ok := v.(map[string]any)
			if !ok {
				return fmt.Errorf("%s must be an object", key)
			}
			for _, v := range m {
				if _, ok := v.(string); !ok {
					return fmt.Errorf("%s values must be strings", key)
				}
			}
		}
	}
	return nil
}
func (s *Service) MCP() ([]MCPServer, []error) {
	byName := map[string]MCPServer{}
	var errs []error
	for _, scope := range []string{"global", "project"} {
		path, e := s.mcpPath(scope)
		if e != nil {
			if s.Workspace != "" {
				errs = append(errs, e)
			}
			continue
		}
		m, e := readObject(path)
		if e != nil {
			errs = append(errs, e)
			continue
		}
		if m["mcpServers"] == nil {
			continue
		}
		servers, ok := m["mcpServers"].(map[string]any)
		if !ok {
			errs = append(errs, fmt.Errorf("%s: mcpServers must be an object", scope))
			continue
		}
		for name, raw := range servers {
			def, _ := raw.(map[string]any)
			if e = ValidateMCP(name, def); e != nil {
				errs = append(errs, fmt.Errorf("%s/%s: %w", scope, name, e))
				continue
			}
			byName[name] = MCPServer{Name: name, Scope: scope, Source: path, Definition: def}
		}
	}
	metadata, metadataError := s.mcpDescriptions()
	if metadataError != nil {
		errs = append(errs, metadataError)
	}
	out := make([]MCPServer, 0, len(byName))
	for _, v := range byName {
		v.Description = metadata[s.mcpMetadataKey(v.Scope, v.Name)]
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, errs
}
func (s *Service) SaveMCP(scope, previous, name string, def map[string]any) error {
	if e := ValidateMCP(name, def); e != nil {
		return e
	}
	path, e := s.mcpPath(scope)
	if e != nil {
		return e
	}
	return s.edit(path, func(m map[string]any) error {
		servers, e := object(m, "mcpServers")
		if e != nil {
			return e
		}
		if previous != name {
			if _, exists := servers[name]; exists {
				return errors.New("a server with that name already exists")
			}
		}
		entry := map[string]any{}
		if old, ok := servers[previous].(map[string]any); ok {
			for k, v := range old {
				entry[k] = v
			}
		}
		for _, key := range []string{"type", "url", "headers", "command", "args", "env", "cwd", "enabled", "exposure", "toolExposure", "timeout", "oauth"} {
			delete(entry, key)
		}
		for k, v := range def {
			entry[k] = v
		}
		if previous != "" && previous != name {
			delete(servers, previous)
		}
		servers[name] = entry
		return nil
	})
}
func (s *Service) RemoveMCP(scope, name string) error {
	path, e := s.mcpPath(scope)
	if e != nil {
		return e
	}
	e = s.edit(path, func(m map[string]any) error {
		servers, e := object(m, "mcpServers")
		if e != nil {
			return e
		}
		delete(servers, name)
		return nil
	})
	if e != nil {
		return e
	}
	metadataError := s.SetMCPDescription(scope, name, "")
	// Only this scope/workspace's generated namespace is eligible. All raw
	// definitions (including globals shadowed by projects) protect references.
	namespace := MCPSecretNamespace(scope, s.Workspace)
	env, secretError := s.SecretEnv()
	if secretError != nil {
		return errors.Join(metadataError, secretError)
	}
	var candidates []string
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, namespace) {
			candidates = append(candidates, key)
		}
	}
	return errors.Join(metadataError, s.PruneMCPSecrets(namespace, candidates))
}
func (s *Service) SetMCPEnabled(scope, name string, enabled bool) error {
	path, e := s.mcpPath(scope)
	if e != nil {
		return e
	}
	return s.edit(path, func(m map[string]any) error {
		servers, e := object(m, "mcpServers")
		if e != nil {
			return e
		}
		entry, ok := servers[name].(map[string]any)
		if !ok {
			return errors.New("server no longer exists")
		}
		entry["enabled"] = enabled
		return nil
	})
}
func (s *Service) ProbeMCP(ctx context.Context) (json.RawMessage, error) {
	out, e := s.Pi(ctx, "mcp", "list", "--json")
	if e != nil {
		var exit *exec.ExitError
		if !errors.As(e, &exit) || exit.ExitCode() != 1 {
			return nil, e
		}
	}
	var m struct {
		Servers []json.RawMessage `json:"servers"`
	}
	if parseError := json.Unmarshal([]byte(out), &m); parseError != nil || m.Servers == nil {
		if e != nil {
			return nil, e
		}
		return nil, errors.New("Pi MCP probe did not return a servers array")
	}
	// Redact decoded strings, not JSON source: multiline/quoted secret values
	// are escaped in JSON and would evade a raw string replacement.
	var doc any
	if e := json.Unmarshal([]byte(out), &doc); e != nil {
		return nil, e
	}
	var scrub func(any) any
	scrub = func(value any) any {
		switch v := value.(type) {
		case string:
			return s.Redact(v)
		case []any:
			for i, item := range v {
				v[i] = scrub(item)
			}
			return v
		case map[string]any:
			for key, item := range v {
				v[key] = scrub(item)
			}
			return v
		default:
			return value
		}
	}
	return json.Marshal(scrub(doc))
}
func (s *Service) SecretEnv() ([]string, error) {
	var m map[string]string
	e := s.Preferences.Read("mcp-secrets.json", &m)
	if e != nil {
		if !errors.Is(e, fs.ErrNotExist) {
			return nil, errors.New("MCP secret store cannot be read; refusing to replace or ignore it")
		}
		return nil, nil
	}
	if m == nil {
		return nil, errors.New("MCP secret store must be an object")
	}
	a := make([]string, 0, len(m))
	for k, v := range m {
		a = append(a, k+"="+v)
	}
	sort.Strings(a)
	return a, nil
}
func (s *Service) SaveSecret(name, value string) error {
	if !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`).MatchString(name) {
		return errors.New("invalid secret variable name")
	}
	return s.Preferences.Update("mcp-secrets.json", func(m map[string]json.RawMessage) error {
		for _, raw := range m {
			var existing string
			if e := json.Unmarshal(raw, &existing); e != nil || string(raw) == "null" {
				return errors.New("MCP secret store must contain only string values; refusing to overwrite it")
			}
		}
		if value == "" {
			delete(m, name)
		} else {
			m[name], _ = json.Marshal(value)
		}
		return nil
	})
}
func (s *Service) Redact(text string) string {
	var m map[string]string
	if s.Preferences.Read("mcp-secrets.json", &m) == nil {
		values := make([]string, 0, len(m))
		for _, v := range m {
			if v != "" {
				values = append(values, v)
			}
		}
		sort.Slice(values, func(i, j int) bool { return len(values[i]) > len(values[j]) })
		for _, v := range values {
			text = strings.ReplaceAll(text, v, "••••••••")
		}
	}
	return bearer.ReplaceAllString(text, "Bearer ••••••••")
}
func SecretReferences(text string) []string {
	var out []string
	for _, m := range secretReference.FindAllStringSubmatch(text, -1) {
		if !slices.Contains(out, m[1]) {
			out = append(out, m[1])
		}
	}
	return out
}
func MCPFingerprint(servers []MCPServer) string {
	b, _ := json.Marshal(servers)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// mcpNumber accepts Go callers as well as numbers decoded from Pi's JSON.
func mcpNumber(value any) (float64, bool) {
	switch n := value.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint64:
		return float64(n), true
	case uint32:
		return float64(n), true
	case json.Number:
		v, e := n.Float64()
		return v, e == nil
	default:
		return 0, false
	}
}
func (s *Service) mcpMetadataKey(scope, name string) string {
	return scope + "|" + s.Workspace + "|" + name
}
func (s *Service) mcpDescriptions() (map[string]string, error) {
	var root map[string]json.RawMessage
	e := s.Preferences.Read("mcp-metadata.json", &root)
	if errors.Is(e, fs.ErrNotExist) {
		return map[string]string{}, nil
	}
	if e != nil {
		return nil, fmt.Errorf("MCP metadata cannot be read: %w", e)
	}
	if root == nil {
		return nil, errors.New("MCP metadata must be an object")
	}
	out := map[string]string{}
	if raw, ok := root["descriptions"]; ok {
		if e = json.Unmarshal(raw, &out); e != nil || out == nil {
			return nil, errors.New("MCP descriptions must be a string map")
		}
	}
	return out, nil
}

// SetMCPDescription writes Orbit-only UX metadata, never Pi's configuration.
func (s *Service) SetMCPDescription(scope, name, description string) error {
	if _, e := s.mcpPath(scope); e != nil {
		return e
	}
	if !mcpName.MatchString(name) {
		return errors.New("invalid server name")
	}
	return s.Preferences.Update("mcp-metadata.json", func(root map[string]json.RawMessage) error {
		descriptions := map[string]string{}
		if raw, ok := root["descriptions"]; ok {
			if e := json.Unmarshal(raw, &descriptions); e != nil || descriptions == nil {
				return errors.New("MCP descriptions must be a string map")
			}
		}
		key := s.mcpMetadataKey(scope, name)
		if strings.TrimSpace(description) == "" {
			delete(descriptions, key)
		} else {
			descriptions[key] = strings.TrimSpace(description)
		}
		root["version"] = json.RawMessage("1")
		root["descriptions"], _ = json.Marshal(descriptions)
		return nil
	})
}

// MCPAutoEnableCodemode resolves Pi's top-level flag with project precedence.
// nil means Pi's default, not a fabricated enabled status.
func (s *Service) MCPAutoEnableCodemode() (*bool, error) {
	var out *bool
	for _, scope := range []string{"global", "project"} {
		if scope == "project" && s.Workspace == "" {
			continue
		}
		path, e := s.mcpPath(scope)
		if e != nil {
			return nil, e
		}
		root, e := readObject(path)
		if e != nil {
			return nil, e
		}
		if raw, ok := root["autoEnableCodemode"]; ok {
			value, ok := raw.(bool)
			if !ok {
				return nil, fmt.Errorf("%s: autoEnableCodemode must be boolean", scope)
			}
			out = &value
		}
	}
	return out, nil
}

// SetMCPAutoEnableCodemode changes only the requested scope's top-level flag.
func (s *Service) SetMCPAutoEnableCodemode(scope string, enabled bool) error {
	path, e := s.mcpPath(scope)
	if e != nil {
		return e
	}
	return s.edit(path, func(root map[string]any) error { root["autoEnableCodemode"] = enabled; return nil })
}

// PruneMCPSecrets deletes only generated candidates from the supplied namespace
// that neither raw global nor raw project configuration references. Read both
// raw files: project shadowing must never hide a live global reference.
func (s *Service) PruneMCPSecrets(namespace string, candidates []string) error {
	if !strings.HasPrefix(namespace, "MCP_") || !strings.HasSuffix(namespace, "_") {
		return errors.New("invalid MCP secret namespace")
	}
	referenced := map[string]bool{}
	for _, scope := range []string{"global", "project"} {
		if scope == "project" && s.Workspace == "" {
			continue
		}
		path, e := s.mcpPath(scope)
		if e != nil {
			return e
		}
		root, e := readObject(path)
		if e != nil {
			return e
		}
		encoded, e := json.Marshal(root)
		if e != nil {
			return e
		}
		for _, name := range SecretReferences(string(encoded)) {
			referenced[name] = true
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	return s.Preferences.Update("mcp-secrets.json", func(root map[string]json.RawMessage) error {
		for _, name := range candidates {
			if strings.HasPrefix(name, namespace) && !referenced[name] {
				delete(root, name)
			}
		}
		return nil
	})
}

// MCPSecretNamespace is source-compatible with Orbit's persisted FNV-1a names.
func MCPSecretNamespace(scope, workspace string) string {
	if scope == "global" {
		return "MCP_GLOBAL_"
	}
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(workspace))
	return fmt.Sprintf("MCP_P%016X_", hash.Sum64())
}
func MCPGeneratedSecretName(server, field, scope, workspace string) string {
	clean := func(value string) string {
		return strings.Map(func(r rune) rune {
			if r >= 'a' && r <= 'z' {
				return r - 'a' + 'A'
			}
			if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
				return r
			}
			return '_'
		}, value)
	}
	return MCPSecretNamespace(scope, workspace) + clean(server) + "_" + clean(field)
}
