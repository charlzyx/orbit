// Package workbench implements the Pi-backed settings pages without UI state.
package workbench

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charlzyx/orbit/internal/preferences"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Service struct {
	Preferences         *preferences.Store
	Workspace, PiBinary string
	mu                  sync.Mutex
}

func New(p *preferences.Store, workspace string) *Service {
	return &Service{Preferences: p, Workspace: workspace, PiBinary: "pi"}
}
func readObject(path string) (map[string]any, error) {
	b, e := os.ReadFile(path)
	if errors.Is(e, fs.ErrNotExist) {
		return map[string]any{}, nil
	}
	if e != nil {
		return nil, e
	}
	if len(bytes.TrimSpace(b)) == 0 {
		return map[string]any{}, nil
	}
	var m map[string]any
	if e = json.Unmarshal(b, &m); e != nil {
		return nil, fmt.Errorf("invalid JSON in %s: %w", path, e)
	}
	if m == nil {
		return nil, fmt.Errorf("%s must be an object", path)
	}
	return m, nil
}
func writeObject(path string, m map[string]any) error {
	b, e := json.MarshalIndent(m, "", "  ")
	if e != nil {
		return e
	}
	return preferences.AtomicWrite(path, append(b, '\n'), 0600)
}
func (s *Service) edit(path string, f func(map[string]any) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, e := readObject(path)
	if e != nil {
		return e
	}
	if e = f(m); e != nil {
		return e
	}
	return writeObject(path, m)
}
func object(m map[string]any, key string) (map[string]any, error) {
	v, ok := m[key]
	if !ok {
		out := map[string]any{}
		m[key] = out
		return out, nil
	}
	out, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an object", key)
	}
	return out, nil
}
func stringValue(v any) string { s, _ := v.(string); return s }
func (s *Service) settingsPath(scope string) (string, error) {
	switch scope {
	case "global", "agent", "user":
		return filepath.Join(s.Preferences.PiDir, "settings.json"), nil
	case "project":
		if s.Workspace == "" {
			return "", errors.New("select a workspace first")
		}
		return filepath.Join(s.Workspace, ".pi", "settings.json"), nil
	default:
		return "", errors.New("invalid configuration scope")
	}
}

// Command executes argv directly, never through a shell. Call from a worker
// goroutine; context cancellation, output limits and child reaping are enforced.
func (s *Service) Command(ctx context.Context, binary string, args ...string) (string, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 4*time.Minute)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = s.Workspace
	cmd.Env = append(os.Environ(), "HOME="+s.Preferences.Home, "PI_CODING_AGENT_DIR="+s.Preferences.PiDir, "PI_SKIP_VERSION_CHECK=1", "NO_COLOR=1", "CI=1")
	secrets, e := s.SecretEnv()
	if e != nil {
		return "", e
	}
	cmd.Env = append(cmd.Env, secrets...)
	cmd.WaitDelay = 2 * time.Second
	var out, errout limitedBuffer
	out.remaining = 16 << 20
	errout.remaining = 1 << 20
	cmd.Stdout = &out
	cmd.Stderr = &errout
	e = cmd.Run()
	if e != nil {
		return "", fmt.Errorf("%s: %w: %s", filepath.Base(binary), e, s.Redact(strings.TrimSpace(errout.String())))
	}
	if out.truncated {
		return "", errors.New("command output exceeded 16 MiB limit")
	}
	return out.String(), nil
}

type limitedBuffer struct {
	bytes.Buffer
	remaining int
	truncated bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if len(p) > b.remaining {
		p = p[:b.remaining]
		b.truncated = true
	}
	b.Buffer.Write(p)
	b.remaining -= len(p)
	return n, nil
}
func (s *Service) Pi(ctx context.Context, args ...string) (string, error) {
	bin := s.PiBinary
	if bin == "" {
		bin = "pi"
	}
	return s.Command(ctx, bin, args...)
}
