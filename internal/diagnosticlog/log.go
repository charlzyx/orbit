// Package diagnosticlog records bounded, local-only application stage diagnostics.
// It deliberately has no API for messages, errors, paths, maps, or arbitrary fields.
package diagnosticlog

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"time"
)

// MaxBytes caps each of app.log and app.log.1. Only these two files are kept.
const MaxBytes int64 = 1 << 20

const (
	currentName  = "app.log"
	previousName = "app.log.1"
)

var (
	ErrUnavailable = errors.New("application diagnostics unavailable")
	ErrHome        = errors.New("application diagnostics require an explicit absolute home")
	versionPattern = regexp.MustCompile(`^v?[0-9]{1,4}\.[0-9]{1,4}\.[0-9]{1,4}(?:-(?:go|mygo|native|experimental|dev|alpha|beta|rc|test)(?:[.-](?:go|mygo|native|experimental|dev|alpha|beta|rc|test|[0-9]{1,10}))*)?$`)
)

// Event is an allowlisted application stage, not user-provided text.
type Event uint8

const (
	EventAppStartup Event = iota + 1
	EventAppShutdown
	EventNativeRun
	EventControllerInit
	EventWorkspaceBrowse
	EventWorkspaceSelect
	EventWorkspaceServices
	EventWorkspacePreferences
	EventWorkspaceSession
	EventWorkspaceFiles
	EventWorkspaceTerminal
	EventWorkspaceGit
	EventWorkspaceGenerators
	EventWorkspaceWorktrees
	EventWorkspaceMCP
	EventWorkspaceResources
	EventWorkspaceWatcher
	EventWorkspaceSettings
	EventRuntimeInstallShutdown
	EventRuntimeAutoUpdateShutdown
)

var eventNames = [...]string{
	"", "app_startup", "app_shutdown", "native_run", "controller_init",
	"workspace_browse", "workspace_select", "workspace_services",
	"workspace_preferences", "workspace_session", "workspace_files",
	"workspace_terminal", "workspace_git", "workspace_generators",
	"workspace_worktrees", "workspace_mcp", "workspace_resources",
	"workspace_watcher", "workspace_settings", "runtime_install_shutdown", "runtime_auto_update_shutdown",
}

// Result describes a stage without including its error or application data.
type Result uint8

const (
	ResultStarted Result = iota + 1
	ResultOK
	ResultCancelled
	ResultError
	ResultSkipped
	ResultBusy
	ResultTimeout
)

var resultNames = [...]string{"", "started", "ok", "cancelled", "error", "skipped", "busy", "timeout"}

type entry struct {
	Timestamp  string `json:"timestamp"`
	RunID      string `json:"run_id"`
	Version    string `json:"version"`
	Event      string `json:"event"`
	Result     string `json:"result"`
	DurationMS int64  `json:"duration_ms"`
}

// Logger serializes direct writes; a completed Record is visible without Close.
// There is no asynchronous queue or userspace buffer to lose during a UI hang.
// Keep one Logger per application process and share it across controllers.
// Filesystem failures disable this logger, never terminate the application.
// The zero value and a nil *Logger are safe disabled loggers.
type Logger struct {
	mu      sync.Mutex
	dir     *os.Root
	file    *os.File
	size    int64
	limit   int64
	runID   string
	version string
}

// Open creates diagnostics under home/.orbit-pi/logs/app.log. home must be an
// explicit absolute preferences home; this package never consults HOME or the
// workspace. It must already exist. App-owned symlinks and unsafe permissions
// are rejected. Returned errors are fixed messages and never contain paths.
// Unknown/non-release version strings are replaced with "unknown".
func Open(home, version string) (*Logger, error) {
	return open(home, version, MaxBytes)
}

func open(home, version string, limit int64) (*Logger, error) {
	if home == "" || !filepath.IsAbs(home) {
		return nil, ErrHome
	}
	h, err := os.OpenRoot(home)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer h.Close()
	prefs, err := subdir(h, ".orbit-pi", false)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer prefs.Close()
	dir, err := subdir(prefs, "logs", true)
	if err != nil {
		return nil, ErrUnavailable
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		dir.Close()
		return nil, ErrUnavailable
	}
	if len(version) > 64 || !versionPattern.MatchString(version) {
		version = "unknown"
	}
	l := &Logger{dir: dir, limit: limit, runID: hex.EncodeToString(id[:]), version: version}
	if err := l.preparePrevious(); err != nil {
		l.disable()
		return nil, ErrUnavailable
	}
	if err := l.openCurrent(); err != nil {
		l.disable()
		return nil, ErrUnavailable
	}
	return l, nil
}

// subdir anchors each app-owned directory before opening children. Lstat and
// inode comparisons reject symlinks and replacements; Root also prevents a
// raced symlink from escaping its parent. The caller-selected home may itself
// use a platform symlink (for example /var -> /private/var on macOS).
func subdir(parent *os.Root, name string, private bool) (*os.Root, error) {
	if err := parent.Mkdir(name, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return nil, err
	}
	before, err := parent.Lstat(name)
	if err != nil || !safeDirectory(before, private) {
		return nil, ErrUnavailable
	}
	dir, err := parent.OpenRoot(name)
	if err != nil {
		return nil, err
	}
	opened, err := dir.Stat(".")
	after, afterErr := parent.Lstat(name)
	if err != nil || afterErr != nil || !safeDirectory(after, private) || !os.SameFile(before, opened) || !os.SameFile(opened, after) {
		dir.Close()
		return nil, ErrUnavailable
	}
	return dir, nil
}

func safeDirectory(info os.FileInfo, private bool) bool {
	if info == nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false
	}
	mask := os.FileMode(0022) // Existing preferences may be readable, never writable by others.
	if private {
		mask = 0077
	}
	return runtime.GOOS == "windows" || info.Mode().Perm()&mask == 0
}

func safeFile(info os.FileInfo) bool {
	return info != nil && info.Mode().IsRegular() && (runtime.GOOS == "windows" || info.Mode().Perm()&0077 == 0)
}

func (l *Logger) preparePrevious() error {
	info, err := l.dir.Lstat(previousName)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil || !safeFile(info) {
		return ErrUnavailable
	}
	if info.Size() > l.limit {
		return l.dir.Remove(previousName)
	}
	return nil
}

func (l *Logger) openCurrent() error {
	return l.openCurrentAttempt(false)
}

func (l *Logger) openCurrentAttempt(replaced bool) error {
	before, err := l.dir.Lstat(currentName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	flags := os.O_RDWR | os.O_APPEND
	if errors.Is(err, os.ErrNotExist) {
		flags |= os.O_CREATE | os.O_EXCL // Never follow a link introduced at creation.
	} else if !safeFile(before) {
		return ErrUnavailable
	}
	f, err := l.dir.OpenFile(currentName, flags, 0600)
	if err != nil {
		return err
	}
	opened, err := f.Stat()
	after, afterErr := l.dir.Lstat(currentName)
	if err != nil || afterErr != nil || !safeFile(opened) || !safeFile(after) || !os.SameFile(opened, after) || (before != nil && !os.SameFile(before, opened)) {
		f.Close()
		return ErrUnavailable
	}
	if opened.Size() > l.limit {
		if replaced {
			f.Close()
			return ErrUnavailable
		}
		// An oversized pre-existing file is unlinked, never truncated through a
		// potentially hard-linked descriptor. Do not retain an oversized backup.
		f.Close()
		if err := l.dir.Remove(currentName); err != nil {
			return err
		}
		return l.openCurrentAttempt(true)
	}
	l.file, l.size = f, opened.Size()
	return nil
}

// Record appends one JSON line synchronously. Invalid enum values or durations
// are dropped. Durations must be nonnegative and no longer than 24 hours.
// A false result means nothing was recorded; callers should continue normally.
// Use ResultStarted immediately before a potentially blocking operation and
// the same Event with a completion result afterwards. Do not log raw errors.
func (l *Logger) Record(event Event, result Result, duration time.Duration) bool {
	if l == nil || event == 0 || int(event) >= len(eventNames) || result == 0 || int(result) >= len(resultNames) || duration < 0 || duration > 24*time.Hour {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return false
	}
	b, err := json.Marshal(entry{Timestamp: time.Now().UTC().Format(time.RFC3339Nano), RunID: l.runID, Version: l.version, Event: eventNames[event], Result: resultNames[result], DurationMS: duration.Milliseconds()})
	if err != nil || int64(len(b)+1) > l.limit {
		return false
	}
	b = append(b, '\n')
	info, err := l.file.Stat()
	current, pathErr := l.dir.Lstat(currentName)
	if err != nil || pathErr != nil || !safeFile(current) || !os.SameFile(info, current) {
		l.disable()
		return false
	}
	l.size = info.Size()
	if l.size > l.limit {
		// Do not rotate externally expanded files into an oversized backup.
		// Disable rather than modifying data this logger did not write.
		l.disable()
		return false
	}
	if l.size+int64(len(b)) > l.limit {
		if err := l.rotate(); err != nil {
			l.disable()
			return false
		}
	}
	n, err := l.file.Write(b)
	if err != nil || n != len(b) {
		l.disable()
		return false
	}
	l.size += int64(n)
	return true
}

func (l *Logger) rotate() error {
	if err := l.preparePrevious(); err != nil {
		return err
	}
	if err := l.file.Close(); err != nil {
		l.file = nil
		return err
	}
	l.file = nil
	if err := l.dir.Remove(previousName); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := l.dir.Rename(currentName, previousName); err != nil {
		return err
	}
	return l.openCurrent()
}

func (l *Logger) disable() {
	if l.file != nil {
		_ = l.file.Close()
		l.file = nil
	}
	if l.dir != nil {
		_ = l.dir.Close()
		l.dir = nil
	}
}

// Close releases the logger. It is nil-safe, idempotent, and safe concurrently
// with Record. Errors are sanitized; records after Close are simply dropped.
func (l *Logger) Close() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	var err error
	if l.file != nil {
		err = l.file.Close()
		l.file = nil
	}
	if l.dir != nil {
		if closeErr := l.dir.Close(); closeErr != nil {
			err = closeErr
		}
		l.dir = nil
	}
	if err != nil {
		return ErrUnavailable
	}
	return nil
}
