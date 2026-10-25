package orbitui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestNativeHeartbeatWaitsForReadyAndInitialWindow(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var ready func()
		stop, done := make(chan struct{}), make(chan struct{})
		initialized, starts := false, 0
		var dispatches atomic.Int32
		registerNativeHostReady(func(f func()) { ready = f }, func() { initialized = true }, func() {
			if !initialized {
				t.Fatal("heartbeat before initial native window setup")
			}
			starts++
			go func() { defer close(done); runHeartbeat(stop, func() bool { dispatches.Add(1); return false }) }()
		})
		time.Sleep(time.Second)
		synctest.Wait()
		if starts != 0 || dispatches.Load() != 0 {
			t.Fatal("owner dispatch before native readiness", starts, dispatches.Load())
		}
		ready()
		synctest.Wait()
		time.Sleep(90 * time.Millisecond)
		synctest.Wait()
		if starts != 1 || dispatches.Load() != 1 {
			t.Fatal("host did not start after native readiness", starts, dispatches.Load())
		}
		// A repeated readiness notification must not create a second retained host.
		// Actual window close/reopen ownership is covered by heartbeat_test.go.
		ready()
		time.Sleep(90 * time.Millisecond)
		synctest.Wait()
		if starts != 1 || dispatches.Load() != 2 {
			t.Fatal("duplicate host lifetime", starts, dispatches.Load())
		}
		close(stop)
		<-done
		time.Sleep(time.Second)
		synctest.Wait()
		if dispatches.Load() != 2 {
			t.Fatal("host survived Quit", dispatches.Load())
		}
	})
}

// The behavioral helper test is coupled to the actual native entrypoint: no
// caller may accidentally restore the old pre-App.Run launch beside it.
func TestNativeRunRegistersHeartbeatInsideReadiness(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "run.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	registrations := 0
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "Run" {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "startHeartbeat" {
				t.Error("Run directly starts heartbeat outside native-ready registration")
			}
			if name, ok := call.Fun.(*ast.Ident); ok && name.Name == "registerNativeHostReady" {
				registrations++
				if len(call.Args) != 3 {
					t.Fatal("invalid readiness registration")
				}
				register, ok := call.Args[0].(*ast.SelectorExpr)
				if !ok || register.Sel.Name != "WhenReady" {
					t.Error("host must use native WhenReady")
				}
				start, ok := call.Args[2].(*ast.SelectorExpr)
				if !ok || start.Sel.Name != "startHeartbeat" {
					t.Error("readiness must own heartbeat launch")
				}
			}
			return true
		})
	}
	if registrations != 1 {
		t.Fatal("expected one native host readiness registration", registrations)
	}
}
