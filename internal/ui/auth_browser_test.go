package orbitui

import (
	"errors"
	"fmt"
	"github.com/charlzyx/orbit/internal/agent"
	"testing"
)

func TestAuthBrowserEffectOpensOnceAndNeverReplaysTerminalOrDeviceState(t *testing.T) {
	x := maintenanceFixture(t)
	calls := []string{}
	open := func(url string) error { calls = append(calls, url); return nil }
	a := agent.AuthState{SessionID: "one", Status: "pending", URL: "https://example.invalid/one"}
	x.openAuthBrowser(a, open)
	a.Status = "waiting"
	for range 3 {
		x.openAuthBrowser(a, open)
	}
	b := a
	b.SessionID = "two"
	b.URL = "https://example.invalid/two"
	x.openAuthBrowser(b, open)
	x.openAuthBrowser(a, open)
	b.Status = "succeeded"
	b.URL = "https://example.invalid/not-opened"
	x.openAuthBrowser(b, open)
	b.Status = "waiting"
	b.URL = ""
	b.VerificationURI = "https://example.invalid/device"
	x.openAuthBrowser(b, open)
	if len(calls) != 2 || calls[0] != a.URL || calls[1] != "https://example.invalid/two" {
		t.Fatal(calls)
	}
}
func TestAuthBrowserFailureDoesNotRetryEachFrameAndLedgerIsBounded(t *testing.T) {
	x := maintenanceFixture(t)
	calls := 0
	a := agent.AuthState{SessionID: "one", Status: "waiting", URL: "https://example.invalid/one"}
	open := func(string) error { calls++; return errors.New("fixture browser unavailable") }
	x.openAuthBrowser(a, open)
	x.openAuthBrowser(a, open)
	if calls != 1 {
		t.Fatal("browser failure retried automatically")
	}
	for i := 0; i < 100; i++ {
		a.SessionID = fmt.Sprint(i)
		x.authBrowsers.take(a)
	}
	if len(x.authBrowsers.seen) != 32 {
		t.Fatal("browser effect ledger unbounded")
	}
}
