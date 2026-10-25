package native

import (
	"net/url"
	"runtime"
	"strings"
	"sync"

	"github.com/egoist/mygo"
)

// Lifecycle parks the application's controller when the last macOS window
// closes. Reopen must create only the view, not another pi session/process.
// Invoke BindLifecycle before App.Run, and SetWindow each time Reopen builds.
type LifecycleOptions struct {
	Reopen      func() *mygo.Window
	OpenSession func(string)
	OpenFile    func(string)
	Quit        func()
	Notifier    *Notifier
}
type Lifecycle struct {
	mu     sync.Mutex
	window *mygo.Window
	opts   LifecycleOptions
	off    []func()
}

func BindLifecycle(opts LifecycleOptions) *Lifecycle {
	l := &Lifecycle{opts: opts}
	if runtime.GOOS == "darwin" {
		l.off = append(l.off, mygo.App.OnWindowAllClosed(func() {}))
	}
	l.off = append(l.off, mygo.App.OnActivate(func(visible bool) {
		if !visible {
			l.Show()
		}
	}))
	l.off = append(l.off, mygo.App.OnNotificationClick(func(id string) {
		if session, ok := SessionFromNotification(id); ok {
			l.Show()
			if opts.OpenSession != nil {
				opts.OpenSession(session)
			}
		}
	}))
	l.off = append(l.off, mygo.App.OnOpenFile(func(path string) {
		l.Show()
		if opts.OpenFile != nil {
			opts.OpenFile(path)
		}
	}))
	l.off = append(l.off, mygo.App.OnOpenURL(func(raw string) {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "orbit" || u.Host != "session" {
			return
		}
		if session := u.Query().Get("path"); session != "" && !strings.ContainsRune(session, 0) {
			l.Show()
			if opts.OpenSession != nil {
				opts.OpenSession(session)
			}
		}
	}))
	l.off = append(l.off, mygo.App.OnQuit(func() {
		if opts.Quit != nil {
			opts.Quit()
		}
	}))
	return l
}
func (l *Lifecycle) SetWindow(w *mygo.Window) {
	l.mu.Lock()
	l.window = w
	l.mu.Unlock()
	if w == nil {
		return
	}
	n := l.opts.Notifier
	if n != nil {
		n.SetForeground(w.IsFocused())
		w.OnFocus(func() { n.SetForeground(true) })
		w.OnBlur(func() { n.SetForeground(false) })
		w.OnClosed(func() { n.SetForeground(false) })
	}
}
func (l *Lifecycle) Window() *mygo.Window { l.mu.Lock(); defer l.mu.Unlock(); return l.window }
func (l *Lifecycle) Show() {
	w := l.Window()
	if w == nil || w.IsDestroyed() {
		if l.opts.Reopen == nil {
			return
		}
		w = l.opts.Reopen()
		l.SetWindow(w)
	}
	if w != nil {
		w.Show()
		w.Focus()
	}
}
func (l *Lifecycle) Close() {
	for _, off := range l.off {
		off()
	}
	l.off = nil
}

// ReduceMotion combines the app's explicit preference with the OS request.
// Call only after App.WhenReady; MyGo's own controls already honor the OS.
func ReduceMotion(appPreference bool) bool {
	return appPreference || mygo.Theme.Preferences().ReduceMotion
}
