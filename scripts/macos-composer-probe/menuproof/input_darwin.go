//go:build darwin && arm64

package menuproof

import (
	"errors"
	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
	"golang.org/x/sys/unix"
	"os"
	"time"
)

type nativeSize struct{ W, H float64 }
type nativeRect struct {
	P point
	S nativeSize
}

var rectMessage func(objc.ID, objc.SEL) nativeRect
var convertRectMessage func(objc.ID, objc.SEL, nativeRect, objc.ID) nativeRect
var convertMessage func(objc.ID, objc.SEL, point) point
var mouseMessage func(objc.ID, objc.SEL, uint, point, uint, float64, int, objc.ID, int, int, float32) objc.ID

func CurrentIdentity(handle uintptr) (Identity, error) {
	if handle == 0 {
		return Identity{}, ErrBlocked
	}
	k, e := unix.SysctlKinfoProc("kern.proc.pid", os.Getpid())
	if e != nil {
		return Identity{}, e
	}
	if int(k.Proc.P_pid) != os.Getpid() || k.Proc.P_starttime.Sec <= 0 {
		return Identity{}, ErrBlocked
	}
	return Identity{os.Getpid(), uint64(k.Proc.P_starttime.Sec)*1_000_000 + uint64(k.Proc.P_starttime.Usec), handle}, nil
}

type Target struct {
	Identity Identity  `json:"identity"`
	At       time.Time `json:"-"`
	AtUnixNS int64     `json:"at_unix_ns"`
	Label    string    `json:"label"`
	Content  Rect      `json:"content"`
	Rect     Rect      `json:"rect"`
	Matches  int       `json:"matches"`
	Nodes    int       `json:"nodes"`
}

func loadInput() error {
	if e := loadNative(); e != nil {
		return e
	}
	if rectMessage == nil {
		purego.RegisterFunc(&rectMessage, native.msg)
		purego.RegisterFunc(&convertMessage, native.msg)
		purego.RegisterFunc(&convertRectMessage, native.msg)
		purego.RegisterFunc(&mouseMessage, native.msg)
	}
	return nil
}
func FreshTarget(id Identity, label string) (Target, error) {
	out := Target{Identity: id, At: time.Now(), Label: label}
	out.AtUnixNS = out.At.UnixNano()
	if e := loadInput(); e != nil {
		return out, e
	}
	now, e := CurrentIdentity(id.Window)
	if e != nil || now != id {
		return out, ErrBlocked
	}
	w := objc.ID(id.Window)
	if send(w, "isVisible") == 0 || send(w, "isKeyWindow") == 0 {
		return out, ErrBlocked
	}
	content := send(w, "contentView")
	bounds := rectMessage(content, objc.RegisterName("bounds"))
	bounds = convertRectMessage(content, objc.RegisterName("convertRect:toView:"), bounds, 0)
	out.Content = Rect{bounds.P.X, bounds.P.Y, bounds.S.W, bounds.S.H}
	if !out.Content.Valid() {
		return out, ErrBlocked
	}
	seen := map[objc.ID]bool{}
	edges := 0
	var walk func(objc.ID, int) error
	walk = func(n objc.ID, depth int) error {
		if n == 0 || seen[n] {
			return nil
		}
		if depth > 30 || out.Nodes >= 2048 {
			return errors.New("AX traversal bound")
		}
		seen[n] = true
		out.Nodes++
		for _, selector := range []string{"isAccessibilityHidden", "isHidden"} {
			if send(n, "respondsToSelector:", uintptr(objc.RegisterName(selector))) != 0 && send(n, selector) != 0 {
				return nil
			}
		}
		if send(n, "respondsToSelector:", uintptr(objc.RegisterName("isAccessibilityElement"))) != 0 && send(n, "isAccessibilityElement") == 0 { /* containers still traversed */
		}
		name := ""
		if send(n, "respondsToSelector:", uintptr(objc.RegisterName("accessibilityLabel"))) != 0 {
			name = text(send(n, "accessibilityLabel"))
		}
		if name == label && send(n, "respondsToSelector:", uintptr(objc.RegisterName("accessibilityFrame"))) != 0 {
			for _, selector := range []string{"isAccessibilityHidden", "isHidden"} {
				if send(n, "respondsToSelector:", uintptr(objc.RegisterName(selector))) != 0 && send(n, selector) != 0 {
					return errors.New("AX target hidden")
				}
			}
			for _, selector := range []string{"isAccessibilityEnabled", "isEnabled"} {
				if send(n, "respondsToSelector:", uintptr(objc.RegisterName(selector))) != 0 && send(n, selector) == 0 {
					return errors.New("AX target disabled")
				}
			}
			r := rectMessage(n, objc.RegisterName("accessibilityFrame"))
			p := convertMessage(w, objc.RegisterName("convertPointFromScreen:"), r.P)
			candidate := Rect{p.X, p.Y, r.S.W, r.S.H}
			if !candidate.Valid() || !out.Content.Contains(candidate) {
				return errors.New("AX target outside finite content bounds")
			}
			out.Matches++
			out.Rect = candidate
		}
		for _, selector := range []string{"accessibilityChildren", "subviews"} {
			if send(n, "respondsToSelector:", uintptr(objc.RegisterName(selector))) == 0 {
				continue
			}
			a := send(n, selector)
			count := int(send(a, "count"))
			if count < 0 || count > 2048 {
				return errors.New("AX child bound")
			}
			for i := 0; i < count; i++ {
				edges++
				if edges > 8192 {
					return errors.New("AX edge bound")
				}
				if e := walk(send(a, "objectAtIndex:", uintptr(i)), depth+1); e != nil {
					return e
				}
			}
		}
		return nil
	}
	if e := walk(content, 0); e != nil {
		return out, e
	}
	if out.Matches != 1 {
		return out, ErrBlocked
	}
	return out, nil
}
func validateTarget(t Target) error {
	now, e := CurrentIdentity(t.Identity.Window)
	if e != nil || now != t.Identity || time.Since(t.At) < 0 || time.Since(t.At) > 250*time.Millisecond || t.Matches != 1 || !t.Content.Contains(t.Rect) {
		return ErrBlocked
	}
	w := objc.ID(now.Window)
	if send(w, "isKeyWindow") == 0 || send(w, "isVisible") == 0 {
		return ErrBlocked
	}
	return nil
}

// PostClick only sends native down/up to this process's already-key window.
// It does not activate another window or use any fallback coordinate.
func PostClick(t Target) error {
	if e := validateTarget(t); e != nil {
		return e
	}
	w := objc.ID(t.Identity.Window)
	app := send(objc.ID(objc.GetClass("NSApplication")), "sharedApplication")
	number := int(send(w, "windowNumber"))
	if number <= 0 {
		return ErrBlocked
	}
	p := point{t.Rect.X + t.Rect.W/2, t.Rect.Y + t.Rect.H/2}
	if app == 0 {
		return ErrBlocked
	}
	return postPair(func(i int) uintptr {
		kind := []uint{1, 2}[i]
		ev := mouseMessage(objc.ID(objc.GetClass("NSEvent")), objc.RegisterName("mouseEventWithType:location:modifierFlags:timestamp:windowNumber:context:eventNumber:clickCount:pressure:"), kind, p, 0, 0, number, 0, 0, 1, 1)
		return uintptr(ev)
	}, func(ev uintptr) { send(app, "postEvent:atStart:", ev, 0) }, nil)
}

// PostKey is deliberately a finite allowlist, not an arbitrary input API.
func PostKey(t Target, key string) error {
	if t.Label != "Message" {
		return ErrBlocked
	}
	if e := validateTarget(t); e != nil {
		return e
	}
	code := uint16(0)
	mods := uint(0)
	chars := key
	switch key {
	case "s":
		code = 1
	case "e":
		code = 14
	case "d":
		code = 2
	case "x":
		code = 7
	case "CmdV":
		code = 9
		mods = 1 << 20
		chars = "v"
	case "CmdZ":
		code = 6
		mods = 1 << 20
		chars = "z"
	default:
		return ErrBlocked
	}
	w := objc.ID(t.Identity.Window)
	number := int(send(w, "windowNumber"))
	if number <= 0 {
		return ErrBlocked
	}
	app := send(objc.ID(objc.GetClass("NSApplication")), "sharedApplication")
	if app == 0 {
		return ErrBlocked
	}
	return postPair(func(i int) uintptr {
		kind := []uint{10, 11}[i]
		ev := native.key(objc.ID(objc.GetClass("NSEvent")), objc.RegisterName("keyEventWithType:location:modifierFlags:timestamp:windowNumber:context:characters:charactersIgnoringModifiers:isARepeat:keyCode:"), kind, point{}, mods, 0, number, 0, str(chars), str(chars), false, code)
		return uintptr(ev)
	}, func(ev uintptr) { send(app, "postEvent:atStart:", ev, 0) }, nil)
}
func ClipboardGeneration() int64 {
	if loadNative() != nil {
		return -1
	}
	p := send(objc.ID(objc.GetClass("NSPasteboard")), "generalPasteboard")
	return int64(send(p, "changeCount"))
}
