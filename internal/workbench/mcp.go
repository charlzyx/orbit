package workbench

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
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
		f, ok := v.(float64)
		if !ok || f <= 0 || f != float64(uint64(f)) {
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
		if oauth, ok := def["oauth"].(map[string]any); ok {
			if port, ok := oauth["callbackPort"].(float64); ok && (port < 1 || port > 65535 || port != float64(int(port))) {
				return errors.New("OAuth callback port must be1–65535")
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
	out := make([]MCPServer, 0, len(byName))
	for _, v := range byName {
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
	return s.edit(path, func(m map[string]any) error {
		servers, e := object(m, "mcpServers")
		if e != nil {
			return e
		}
		delete(servers, name)
		return nil
	})
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
	out, e := s.Pi(ctx, "mcp", "list", "--json", "--approve")
	if e != nil {
		return nil, e
	}
	var m struct {
		Servers []json.RawMessage `json:"servers"`
	}
	if e = json.Unmarshal([]byte(out), &m); e != nil || m.Servers == nil {
		return nil, errors.New("Pi MCP probe did not return a servers array")
	}
	return json.RawMessage(s.Redact(out)), nil
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
	a := make([]string, 0, len(m))
	for k, v := range m {
		a = append(a, k+"="+v)
	}
	return a, nil
}
func (s *Service) SaveSecret(name, value string) error {
	if !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`).MatchString(name) {
		return errors.New("invalid secret variable name")
	}
	return s.Preferences.Update("mcp-secrets.json", func(m map[string]json.RawMessage) error {
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
