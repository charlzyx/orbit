//go:build darwin && arm64

package menuproof

import (
	"errors"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// NativeObserver is test-only and unwired. Every method/callback must run on
// the app's main thread. It never pumps a replacement event loop or mutates a
// menu delegate. OwnIdentity must read the real current app/window identity.
type NativeObserver struct {
	State          State
	OwnIdentity    func() Identity
	OnEscapePosted func()
	object         objc.ID
	center         objc.ID
	closed         bool
}

var native struct {
	once          sync.Once
	err           error
	msg           uintptr
	class         objc.Class
	observers     map[objc.ID]*NativeObserver
	notifications [3]objc.ID
	key           func(objc.ID, objc.SEL, uint, point, uint, float64, int, objc.ID, objc.ID, objc.ID, bool, uint16) objc.ID
	number        func(objc.ID, objc.SEL) float64
}

type point struct{ X, Y float64 }

//go:uintptrescapes
func send(o objc.ID, sel string, args ...uintptr) objc.ID {
	if o == 0 {
		return 0
	}
	var a [12]uintptr
	a[0] = uintptr(o)
	a[1] = uintptr(objc.RegisterName(sel))
	n := copy(a[2:], args)
	r, _, _ := purego.SyscallN(native.msg, a[:n+2]...)
	return objc.ID(r)
}
func str(s string) objc.ID {
	p := unsafe.StringData(s)
	o := send(send(objc.ID(objc.GetClass("NSString")), "alloc"), "initWithBytes:length:encoding:", uintptr(unsafe.Pointer(p)), uintptr(len(s)), 4)
	runtime.KeepAlive(s)
	return send(o, "autorelease")
}
func text(o objc.ID) string {
	n := send(o, "lengthOfBytesUsingEncoding:", 4)
	if n == 0 || n > 320 {
		return ""
	}
	p := send(o, "UTF8String")
	if p == 0 {
		return ""
	}
	return string(unsafe.Slice(*(**byte)(unsafe.Pointer(&p)), int(n)))
}
func loadNative() error {
	native.once.Do(func() {
		lib, e := purego.Dlopen("/usr/lib/libobjc.A.dylib", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if e != nil {
			native.err = e
			return
		}
		native.msg, e = purego.Dlsym(lib, "objc_msgSend")
		if e != nil {
			native.err = e
			return
		}
		purego.RegisterFunc(&native.key, native.msg)
		purego.RegisterFunc(&native.number, native.msg)
		appkit, e := purego.Dlopen("/System/Library/Frameworks/AppKit.framework/AppKit", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if e != nil {
			native.err = e
			return
		}
		for i, name := range []string{"NSMenuDidBeginTrackingNotification", "NSMenuDidEndTrackingNotification", "NSMenuWillSendActionNotification"} {
			p, e := purego.Dlsym(appkit, name)
			if e != nil {
				native.err = e
				return
			}
			// Dlsym returns a C-owned static data-symbol address, not a Go
			// allocation. Reinterpret its pointer-sized slot before dereferencing
			// the exported NSString*. The AppKit handle stays loaded.
			symbolPointer := *(*unsafe.Pointer)(unsafe.Pointer(&p))
			native.notifications[i] = *(*objc.ID)(symbolPointer)
			if native.notifications[i] == 0 {
				native.err = errors.New("missing public menu notification")
				return
			}
		}
		native.observers = map[objc.ID]*NativeObserver{}
		native.class, e = objc.RegisterClass("OrbitComposerProofMenuObserver", objc.GetClass("NSObject"), nil, nil, []objc.MethodDef{
			{Cmd: objc.RegisterName("orbitMenuBegin:"), Fn: func(self objc.ID, _ objc.SEL, n objc.ID) {
				if o := native.observers[self]; o != nil {
					o.begin(send(n, "object"))
				}
			}},
			{Cmd: objc.RegisterName("orbitMenuEnd:"), Fn: func(self objc.ID, _ objc.SEL, n objc.ID) {
				if o := native.observers[self]; o != nil {
					o.State.End(o.observe(send(n, "object"), false))
				}
			}},
			{Cmd: objc.RegisterName("orbitMenuAction:"), Fn: func(self objc.ID, _ objc.SEL, n objc.ID) {
				if o := native.observers[self]; o != nil {
					o.State.Action(uintptr(send(n, "object")))
				}
			}},
		})
		native.err = e
	})
	return native.err
}

func NewNativeObserver(currentIdentity func() Identity) (*NativeObserver, error) {
	if currentIdentity == nil {
		return nil, ErrBlocked
	}
	if e := loadNative(); e != nil {
		return nil, e
	}
	if len(native.observers) != 0 {
		return nil, ErrBlocked
	}
	o := &NativeObserver{OwnIdentity: currentIdentity}
	o.object = send(send(objc.ID(native.class), "alloc"), "init")
	if o.object == 0 {
		return nil, ErrBlocked
	}
	native.observers[o.object] = o
	o.center = send(objc.ID(objc.GetClass("NSNotificationCenter")), "defaultCenter")
	for i, selector := range []string{"orbitMenuBegin:", "orbitMenuEnd:", "orbitMenuAction:"} {
		send(o.center, "addObserver:selector:name:object:", uintptr(o.object), uintptr(objc.RegisterName(selector)), uintptr(native.notifications[i]), 0)
	}
	return o, nil
}
func (o *NativeObserver) observe(menu objc.ID, labels bool) Observation {
	id := o.OwnIdentity()
	win := objc.ID(id.Window)
	v := Observation{Owner: id, Menu: uintptr(menu), At: time.Now(), Key: send(win, "isKeyWindow") != 0, Visible: send(win, "isVisible") != 0}
	if labels && send(menu, "isKindOfClass:", uintptr(objc.GetClass("NSMenu"))) != 0 && send(menu, "numberOfItems") == 4 {
		for i := 0; i < 4; i++ {
			v.Labels = append(v.Labels, text(send(send(menu, "itemAtIndex:", uintptr(i)), "title")))
		}
	}
	return v
}
func (o *NativeObserver) begin(menu objc.ID) {
	if o.closed || !o.State.Begin(o.observe(menu, true)) {
		return
	}
	// Notification callback runs within the real popup path. Queue once here,
	// not through a main-thread task which may wait behind the menu tracking loop.
	// This is ordinary native Escape dispatch, never cancelTracking/performAction.
	win := objc.ID(o.State.arm.Owner.Window)
	number := int(send(win, "windowNumber"))
	if number <= 0 {
		o.State.Failed = "invalid native window number"
		return
	}
	app := send(objc.ID(objc.GetClass("NSApplication")), "sharedApplication")
	clock := send(objc.ID(objc.GetClass("NSProcessInfo")), "processInfo")
	stamp := native.number(clock, objc.RegisterName("systemUptime"))
	chars := str("\x1b")
	if app == 0 {
		o.State.Failed = "native application unavailable"
		return
	}
	if err := postPair(func(i int) uintptr {
		kind := []uint{10, 11}[i]
		ev := native.key(objc.ID(objc.GetClass("NSEvent")), objc.RegisterName("keyEventWithType:location:modifierFlags:timestamp:windowNumber:context:characters:charactersIgnoringModifiers:isARepeat:keyCode:"), kind, point{}, 0, stamp, number, 0, chars, chars, false, 53)
		return uintptr(ev)
	}, func(ev uintptr) { send(app, "postEvent:atStart:", ev, 0) }, o.OnEscapePosted); err != nil {
		o.State.Failed = "native Escape pair construction failed"
	}
}
func (o *NativeObserver) Close() {
	if o == nil || o.closed {
		return
	}
	o.closed = true
	send(o.center, "removeObserver:", uintptr(o.object))
	delete(native.observers, o.object)
	send(o.object, "release")
	o.object = 0
	o.State.Close()
	o.OwnIdentity = nil
}
