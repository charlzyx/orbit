// Package native connects Orbit's Go workbench to real desktop services.
package native

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/egoist/mygo"
)

const SessionNotificationPrefix = "orbit-session:"
const BodyPreviewCharacters = 180

type NotificationPrefs struct {
	Desktop bool `json:"desktop"`
	Sound   bool `json:"sound"`
	Toasts  bool `json:"toasts"`
}

func DefaultNotificationPrefs() NotificationPrefs { return NotificationPrefs{true, true, true} }

// Missing or malformed individual fields keep their defaults, rather than
// unexpectedly silencing notifications while migrating existing preferences.
func LoadNotificationPrefs(path string) (NotificationPrefs, error) {
	p := DefaultNotificationPrefs()
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return p, nil
	}
	if err != nil {
		return p, err
	}
	var v map[string]json.RawMessage
	if err = json.Unmarshal(raw, &v); err != nil {
		return p, err
	}
	for name, field := range map[string]*bool{"desktop": &p.Desktop, "sound": &p.Sound, "toasts": &p.Toasts} {
		var b bool
		if value, ok := v[name]; ok && json.Unmarshal(value, &b) == nil && string(value) != "null" {
			*field = b
		}
	}
	return p, nil
}
func (p NotificationPrefs) Save(path string) error { return atomicJSON(path, p) }
func PreferencePath(name string) string {
	home := os.Getenv("ORBIT_PREFERENCES_HOME")
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	return filepath.Join(home, ".orbit-pi", name)
}
func atomicJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".orbit-prefs-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(append(raw, '\n'))
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Preview matches Orbit's whitespace/control cleanup and Unicode-safe word
// truncation. The ellipsis is extra, as in the original implementation.
func Preview(text string, limit int) string {
	if limit < 0 {
		limit = 0
	}
	words := strings.Fields(text)
	for i, w := range words {
		words[i] = strings.Map(func(r rune) rune {
			if unicode.IsControl(r) {
				return -1
			}
			return r
		}, w)
	}
	flat := strings.Join(words, " ")
	r := []rune(flat)
	if len(r) <= limit {
		return flat
	}
	end := limit
	for i := limit - 1; i >= limit/2; i-- {
		if r[i] == ' ' {
			end = i
			break
		}
	}
	return strings.TrimSpace(string(r[:end])) + "…"
}

type ToastKind string

const (
	ToastInfo    ToastKind = "info"
	ToastSuccess ToastKind = "success"
	ToastWarning ToastKind = "warning"
	ToastError   ToastKind = "error"
)

func (k ToastKind) TTL() time.Duration {
	switch k {
	case ToastSuccess:
		return 3 * time.Second
	case ToastWarning:
		return 6 * time.Second
	case ToastError:
		return 8 * time.Second
	default:
		return 4 * time.Second
	}
}

const ToastFade = 180 * time.Millisecond
const MaximumToasts = 4

type Toast struct {
	ID          uint64
	Kind        ToastKind
	Title, Body string
	Created     time.Time
	Revision    uint64
}
type ToastStack struct {
	mu     sync.Mutex
	nextID uint64
	items  []Toast
}

func (s *ToastStack) Push(kind ToastKind, title, body string) uint64 {
	return s.PushAt(kind, title, body, time.Now())
}
func (s *ToastStack) PushAt(kind ToastKind, title, body string, now time.Time) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		v := &s.items[i]
		if v.Kind == kind && v.Title == title && v.Body == body {
			v.Created = now
			v.Revision++
			return v.ID
		}
	}
	id := s.nextID
	s.nextID++
	if len(s.items) >= MaximumToasts {
		s.items = s.items[1:]
	}
	s.items = append(s.items, Toast{ID: id, Kind: kind, Title: title, Body: body, Created: now})
	return id
}
func (s *ToastStack) Items() []Toast {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Toast(nil), s.items...)
}
func (s *ToastStack) Dismiss(id uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, v := range s.items {
		if v.ID == id {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return true
		}
	}
	return false
}
func (s *ToastStack) Expire(now time.Time, reduceMotion bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	fade := ToastFade
	if reduceMotion {
		fade = 0
	}
	n := len(s.items)
	keep := s.items[:0]
	for _, v := range s.items {
		if now.Sub(v.Created) < v.Kind.TTL()+fade {
			keep = append(keep, v)
		}
	}
	s.items = keep
	return len(keep) != n
}
func (t Toast) Opacity(now time.Time, reduceMotion bool) float32 {
	age := now.Sub(t.Created) - t.Kind.TTL()
	if age <= 0 {
		return 1
	}
	if reduceMotion {
		return 0
	}
	return max(0, 1-float32(age)/float32(ToastFade))
}

type Notification struct {
	Session, Title, Subtitle, Body string
	Kind                           ToastKind
}
type NotificationDelivery struct {
	Banner, Sound, Toast bool
	Error                error
}

// Notifier's hooks are on the caller's goroutine; notification I/O should be
// called off the UI thread. BindLifecycle routes desktop clicks on main.
type Notifier struct {
	mu         sync.Mutex
	prefs      NotificationPrefs
	foreground bool
	OnToast    func(ToastKind, string, string)
	ShowBanner func(mygo.NotificationOptions) error
	PlaySound  func()
}

func NewNotifier(prefs NotificationPrefs) *Notifier {
	return &Notifier{prefs: prefs, ShowBanner: func(o mygo.NotificationOptions) error { return showDesktopNotification(o) }, PlaySound: mygo.Shell.Beep}
}
func (n *Notifier) SetPrefs(p NotificationPrefs) { n.mu.Lock(); n.prefs = p; n.mu.Unlock() }
func (n *Notifier) Prefs() NotificationPrefs     { n.mu.Lock(); defer n.mu.Unlock(); return n.prefs }
func (n *Notifier) SetForeground(active bool)    { n.mu.Lock(); n.foreground = active; n.mu.Unlock() }
func (n *Notifier) Deliver(event Notification) NotificationDelivery {
	n.mu.Lock()
	p, front := n.prefs, n.foreground
	n.mu.Unlock()
	result := NotificationDelivery{}
	title := event.Title
	if strings.TrimSpace(title) == "" {
		title = "Orbit Pi"
	}
	title = Preview(title, 64)
	body := Preview(event.Body, BodyPreviewCharacters)
	if front {
		if p.Toasts && n.OnToast != nil {
			n.OnToast(event.Kind, title, body)
			result.Toast = true
		}
		return result
	}
	if p.Desktop && n.ShowBanner != nil {
		id := ""
		if event.Session != "" {
			id = SessionNotificationPrefix + event.Session
		}
		result.Error = n.ShowBanner(mygo.NotificationOptions{ID: id, Title: title, Subtitle: Preview(event.Subtitle, 64), Body: body, Silent: true, Group: event.Session})
		result.Banner = result.Error == nil
	}
	if p.Sound && n.PlaySound != nil {
		n.PlaySound()
		result.Sound = true
	}
	return result
}
func SessionFromNotification(id string) (string, bool) {
	s, ok := strings.CutPrefix(id, SessionNotificationPrefix)
	return s, ok && s != "" && !strings.ContainsRune(s, 0)
}

// DesktopNotificationMode is honest about the unbundled macOS fallback: its
// banner belongs to Script Editor and cannot route a click back to a session.
type DesktopNotificationMode string

const (
	NotificationBundled     DesktopNotificationMode = "bundled"
	NotificationUnbundled   DesktopNotificationMode = "unbundled"
	NotificationUnavailable DesktopNotificationMode = "unavailable"
)

func NotificationMode() DesktopNotificationMode {
	if runtime.GOOS == "darwin" && !mygo.App.IsPackaged() {
		return NotificationUnbundled
	}
	if mygo.NotificationsSupported() {
		return NotificationBundled
	}
	return NotificationUnavailable
}
func showDesktopNotification(o mygo.NotificationOptions) error {
	if runtime.GOOS != "darwin" || mygo.App.IsPackaged() {
		return mygo.NewNotification(o).Show()
	}
	const script = "on run argv\n display notification (item 3 of argv) with title (item 1 of argv) subtitle (item 2 of argv)\nend run"
	// Text is passed as argv, not concatenated into AppleScript source.
	cmd := exec.Command("/usr/bin/osascript", "-e", script, "--", o.Title, o.Subtitle, o.Body)
	return cmd.Run()
}
