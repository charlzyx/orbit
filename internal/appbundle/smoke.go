package appbundle

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type SmokeChecks struct {
	NativeCapture     bool `json:"native_capture"`
	NativeWindowReady bool `json:"native_window_ready"`
	FirstFrame        bool `json:"first_frame"`
	GhosttyLoad       bool `json:"ghostty_load"`
	PiRPCGetState     bool `json:"pi_rpc_get_state"`
}
type SmokeReport struct {
	CapturePath   string      `json:"capture_path,omitempty"`
	CaptureWidth  int         `json:"capture_width,omitempty"`
	CaptureHeight int         `json:"capture_height,omitempty"`
	CaptureSHA256 string      `json:"capture_sha256,omitempty"`
	CaptureKind   string      `json:"capture_kind,omitempty"`
	Status        string      `json:"status"`
	PID           int         `json:"pid"`
	Architecture  string      `json:"architecture"`
	StartedAt     time.Time   `json:"started_at"`
	Executable    string      `json:"executable"`
	Home          string      `json:"home"`
	UserHome      string      `json:"user_home"`
	PiAgentDir    string      `json:"pi_agent_dir"`
	Workspace     string      `json:"workspace"`
	GhosttyPath   string      `json:"ghostty_path"`
	PiPath        string      `json:"pi_path"`
	BundledPiPath string      `json:"bundled_pi_path"`
	NodePath      string      `json:"node_path"`
	PiPID         int         `json:"pi_pid,omitempty"`
	SessionID     string      `json:"session_id,omitempty"`
	RuntimeID     string      `json:"runtime_id,omitempty"`
	Checks        SmokeChecks `json:"checks"`
	Error         string      `json:"error,omitempty"`
}

type SmokeSnapshot struct {
	Connected, Alive, Loading      bool
	PID                            int
	SessionID, RuntimeID, Launcher string
}

// Smoke is opt-in CI instrumentation. It never synthesizes app state or sends
// prompts. PASS requires the real window's first surface frame and the active
// manager's successful Pi bootstrap, which includes get_state.
type Smoke struct {
	path          string
	timeout       time.Duration
	ready         atomic.Bool
	mu            sync.Mutex
	report        SmokeReport
	written       bool
	captureOnce   sync.Once
	captureFailed bool
}

func NewSmoke(path string, timeout time.Duration, l Layout, p Prepared) (*Smoke, error) {
	if path == "" || !filepath.IsAbs(path) {
		return nil, errors.New("--smoke-report must be an absolute path")
	}
	if timeout <= 0 || timeout > 5*time.Minute {
		return nil, errors.New("--smoke-exit-after must be greater than zero and at most 5m")
	}
	if err := l.checkWritableLocation(path); err != nil {
		return nil, err
	}
	if _, err := os.Lstat(path); err == nil || !errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("Smoke report path already exists or cannot be inspected; choose a fresh path")
	}
	if _, err := os.Lstat(path + ".png"); err == nil || !errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("Smoke capture path already exists or cannot be inspected; choose a fresh path")
	}
	return &Smoke{path: path, timeout: timeout, report: SmokeReport{Status: "FAIL", PID: os.Getpid(), Architecture: runtime.GOARCH, StartedAt: time.Now().UTC(), Executable: l.Executable, Home: p.Home, UserHome: p.UserHome, PiAgentDir: p.PiAgentDir, Workspace: p.Workspace, GhosttyPath: l.Ghostty, PiPath: p.PiBinary, BundledPiPath: l.Pi, NodePath: l.Node}}, nil
}

// WindowReady is called only by MyGo's OnReadyToShow after its first surface
// frame, and only after the actual native window reports visible.
func (s *Smoke) WindowReady(visible bool) {
	if s != nil && visible {
		s.ready.Store(true)
	}
}
func (s *Smoke) GhosttyLoaded() {
	if s != nil {
		s.mu.Lock()
		s.report.Checks.GhosttyLoad = true
		s.mu.Unlock()
	}
}

func (s *Smoke) Start(ctx context.Context, snapshot func() SmokeSnapshot, quit func()) {
	if s == nil {
		return
	}
	go func() {
		deadline := time.NewTimer(s.timeout)
		defer deadline.Stop()
		tick := time.NewTicker(50 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-deadline.C:
				_ = s.Fail("Startup timed out before all native-window, Ghostty and Pi RPC checks passed")
				quit()
				return
			case <-tick.C:
				current := snapshot()
				s.mu.Lock()
				s.report.Checks.NativeWindowReady = s.ready.Load()
				s.report.Checks.FirstFrame = s.ready.Load()
				// Connected alone is insufficient: a restored local session can carry an
				// identity. Require the live, fully bootstrapped process of this launch.
				s.report.Checks.PiRPCGetState = current.Connected && current.Alive && !current.Loading && current.PID > 0 && current.SessionID != "" && current.Launcher == s.report.PiPath
				checks := s.report.Checks
				pass := checks.NativeWindowReady && checks.FirstFrame && checks.GhosttyLoad && checks.PiRPCGetState && checks.NativeCapture
				failed := s.captureFailed
				if pass {
					s.report.Status = "PASS"
					s.report.PiPID = current.PID
					s.report.SessionID = current.SessionID
					s.report.RuntimeID = current.RuntimeID
				}
				s.mu.Unlock()
				if pass || failed {
					_ = s.write()
					quit()
					return
				}
			}
		}
	}()
}

func (s *Smoke) Fail(message string) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	s.report.Status = "FAIL"
	s.report.Error = message
	s.mu.Unlock()
	return s.write()
}
func (s *Smoke) write() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.written {
		return nil
	}
	data, err := json.MarshalIndent(s.report, "", "  ")
	if err != nil {
		return err
	}
	// O_EXCL prevents a stale success or another process's report being replaced.
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("Cannot write startup smoke report: %w", err)
	}
	_, err = file.Write(append(data, '\n'))
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		s.written = true
	}
	return err
}

// Capture records the actual native content from the running window. MyGo's
// native Content capture renders the current content, not an OS desktop grab;
// first-frame/visibility evidence comes separately from WindowReady.
func (s *Smoke) Capture(capture func() ([]byte, error)) {
	if s == nil {
		return
	}
	s.captureOnce.Do(func() {
		go func() {
			data, err := capture()
			width, height := 0, 0
			if err == nil {
				var configErr error
				config, configErr := png.DecodeConfig(bytes.NewReader(data))
				err = configErr
				width, height = config.Width, config.Height
			}
			if err == nil && (width <= 0 || height <= 0) {
				err = errors.New("Native capture is empty")
			}
			path := s.path + ".png"
			if err == nil {
				var file *os.File
				file, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
				if err == nil {
					_, err = file.Write(data)
					if closeErr := file.Close(); err == nil {
						err = closeErr
					}
				}
			}
			s.mu.Lock()
			defer s.mu.Unlock()
			if err != nil {
				s.captureFailed = true
				s.report.Error = "Native content capture failed"
				return
			}
			hash := sha256.Sum256(data)
			s.report.CapturePath = path
			s.report.CaptureWidth = width
			s.report.CaptureHeight = height
			s.report.CaptureSHA256 = hex.EncodeToString(hash[:])
			s.report.CaptureKind = "native-content"
			s.report.Checks.NativeCapture = true
		}()
	})
}

// SmokeEnvironment keeps opt-in native CI instrumentation credential-free even
// if LaunchServices supplies extra GUI-session variables. Normal user launches
// do not use this allowlist and inherit their ordinary Pi environment.
func SmokeEnvironment(ambient []string) []string {
	var env []string
	for _, entry := range ambient {
		key, _, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		switch key {
		case "HOME", "USERPROFILE", "PATH", "TMPDIR", "LANG", "LC_ALL", "LC_CTYPE", "TZ", "__CF_USER_TEXT_ENCODING", "PI_CODING_AGENT_DIR":
			env = append(env, entry)
		}
	}
	return append(env, "PI_OFFLINE=1", "PI_SKIP_VERSION_CHECK=1")
}
