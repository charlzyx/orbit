//go:build darwin && arm64

package menuproof

import (
	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
	"unsafe"
)

// WindowProof is metadata for exactly one caller-specified window. No titles,
// other windows, pixels, or permission requests are collected here.
type WindowProof struct {
	Process
	Number    uint32 `json:"window_number"`
	Onscreen  bool   `json:"onscreen"`
	Preflight bool   `json:"preflight"`
	Bounds    Rect   `json:"bounds"`
}

func InspectWindow(pid int, start uint64, number uint32) (WindowProof, error) {
	var out WindowProof
	if number == 0 || start == 0 {
		return out, ErrBlocked
	}
	before, err := InspectProcess(pid)
	if err != nil || before.Start != start {
		return out, ErrBlocked
	}
	cg, err := purego.Dlopen("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics", purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		return out, err
	}
	defer purego.Dlclose(cg)
	cf, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		return out, err
	}
	defer purego.Dlclose(cf)
	bind := func(lib uintptr, name string, fn any) error {
		p, e := purego.Dlsym(lib, name)
		if e != nil {
			return e
		}
		purego.RegisterFunc(fn, p)
		return nil
	}
	var preflight func() bool
	var query func(uint32, uint32) uintptr
	var release func(uintptr)
	var count func(uintptr) int64
	var at func(uintptr, int64) uintptr
	var get func(uintptr, uintptr) uintptr
	var numberValue func(uintptr, int64, *int64) bool
	var boolValue func(uintptr) bool
	var rectValue func(uintptr, *nativeRect) bool
	var typeID func(uintptr) uint64
	var arrayType, dictType, numberType, boolType func() uint64
	for _, b := range []struct {
		lib  uintptr
		name string
		fn   any
	}{
		{cg, "CGPreflightScreenCaptureAccess", &preflight}, {cg, "CGWindowListCopyWindowInfo", &query}, {cg, "CGRectMakeWithDictionaryRepresentation", &rectValue},
		{cf, "CFRelease", &release}, {cf, "CFArrayGetCount", &count}, {cf, "CFArrayGetValueAtIndex", &at}, {cf, "CFDictionaryGetValue", &get}, {cf, "CFNumberGetValue", &numberValue}, {cf, "CFBooleanGetValue", &boolValue}, {cf, "CFGetTypeID", &typeID}, {cf, "CFArrayGetTypeID", &arrayType}, {cf, "CFDictionaryGetTypeID", &dictType}, {cf, "CFNumberGetTypeID", &numberType}, {cf, "CFBooleanGetTypeID", &boolType},
	} {
		if err = bind(b.lib, b.name, b.fn); err != nil {
			return out, err
		}
	}
	if !preflight() {
		return out, ErrBlocked
	}
	// kCGWindowListOptionIncludingWindow alone: NEVER all/on-screen enumeration.
	list := query(1<<3, number)
	if list == 0 {
		return out, ErrBlocked
	}
	defer release(list)
	if typeID(list) != arrayType() || count(list) != 1 {
		return out, ErrBlocked
	}
	dict := at(list, 0)
	if dict == 0 || typeID(dict) != dictType() {
		return out, ErrBlocked
	}
	value := func(name string) uintptr {
		p, e := purego.Dlsym(cg, name)
		if e != nil {
			return 0
		}
		slot := *(*unsafe.Pointer)(unsafe.Pointer(&p))
		key := *(*uintptr)(slot)
		return get(dict, key)
	}
	integer := func(name string) (int64, bool) {
		v := value(name)
		var n int64
		if v == 0 || typeID(v) != numberType() {
			return 0, false
		}
		ok := numberValue(v, 4, &n)
		return n, ok
	}
	n, ok := integer("kCGWindowNumber")
	if !ok || n != int64(number) {
		return out, ErrBlocked
	}
	owner, ok := integer("kCGWindowOwnerPID")
	if !ok || owner != int64(pid) {
		return out, ErrBlocked
	}
	visible := value("kCGWindowIsOnscreen")
	if visible == 0 || typeID(visible) != boolType() || !boolValue(visible) {
		return out, ErrBlocked
	}
	bounds := value("kCGWindowBounds")
	var rect nativeRect
	if bounds == 0 || typeID(bounds) != dictType() || !rectValue(bounds, &rect) {
		return out, ErrBlocked
	}
	r := Rect{rect.P.X, rect.P.Y, rect.S.W, rect.S.H}
	if !r.Valid() {
		return out, ErrBlocked
	}
	after, err := InspectProcess(pid)
	if err != nil || after != before {
		return out, ErrBlocked
	}
	return WindowProof{before, number, true, true, r}, nil
}

// OwnWindowReceipt must run on the owned app's main thread. It is an unwired
// phase-observer primitive, not permission to capture or send native input.
func OwnWindowReceipt(id Identity) (uint32, Rect, error) {
	if err := loadInput(); err != nil {
		return 0, Rect{}, err
	}
	current, err := CurrentIdentity(id.Window)
	if err != nil || current != id {
		return 0, Rect{}, ErrBlocked
	}
	w := objc.ID(id.Window)
	n := uint64(send(w, "windowNumber"))
	if n == 0 || n > 0xffffffff || send(w, "isVisible") == 0 {
		return 0, Rect{}, ErrBlocked
	}
	frame := rectMessage(w, objc.RegisterName("frame"))
	r := Rect{frame.P.X, frame.P.Y, frame.S.W, frame.S.H}
	if !r.Valid() {
		return 0, Rect{}, ErrBlocked
	}
	current, err = CurrentIdentity(id.Window)
	if err != nil || current != id {
		return 0, Rect{}, ErrBlocked
	}
	return uint32(n), r, nil
}
