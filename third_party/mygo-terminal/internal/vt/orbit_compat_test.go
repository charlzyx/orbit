package vt

import (
	"testing"
	"unsafe"
)

func TestOrbitLineBudgetUsesVerifiedNativeOption(t *testing.T) {
	loadLib(t)
	term, err := NewTerminal(80, 24, Effects{})
	if err != nil {
		t.Fatal(err)
	}
	defer term.Free()
	if err := term.SetScrollbackLines(10000); err != nil {
		t.Fatal(err)
	}
	var lines, bytes uintptr
	// GhosttyTerminalData values from the same pinned public C header.
	if err := term.get(35, unsafe.Pointer(&lines)); err != nil || lines != 10000 {
		t.Fatalf("native line budget = %d, %v", lines, err)
	}
	if err := term.get(34, unsafe.Pointer(&bytes)); err == nil {
		t.Fatalf("byte budget still limits the line budget: %d", bytes)
	}
}
