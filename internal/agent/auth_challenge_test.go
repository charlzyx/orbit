package agent

import (
	"errors"
	"testing"
)

func TestNewAuthStartClearsPreviousChallengeAndRejectsStaleUpdates(t *testing.T) {
	s := newState("", "")
	s.snapshot.Auth = AuthState{Support: "supported", Providers: []AuthProvider{{ID: "p", Authenticated: true}}, SessionID: "old", Provider: "old-p", Method: "old", Status: "failed", URL: "https://example.invalid/old", UserCode: "OLD-FIXTURE", VerificationURI: "https://example.invalid/old-device", ExpiresAt: 123, Error: "old error"}
	apply(t, &s, `{"type":"auth_login_started","sessionId":"new","provider":"p","method":"browser"}`)
	a := s.snapshot.Auth
	if a.SessionID != "new" || a.Provider != "p" || a.Status != "pending" || a.Error != "" || a.URL != "" || a.UserCode != "" || a.VerificationURI != "" || a.ExpiresAt != 0 {
		t.Fatalf("new start reused challenge: %+v", a)
	}
	if a.Support != "supported" || len(a.Providers) != 1 || !a.Providers[0].Authenticated {
		t.Fatal("new challenge discarded provider state")
	}
	apply(t, &s, `{"type":"auth_device_code","sessionId":"old","userCode":"STALE","verificationUri":"https://example.invalid/stale"}`)
	if s.snapshot.Auth.UserCode != "" {
		t.Fatal("old device challenge entered new login")
	}
	apply(t, &s, `{"type":"auth_device_code","sessionId":"new","userCode":"NEW-FIXTURE","verificationUri":"https://example.invalid/new","expiresAt":456}`)
	if s.snapshot.Auth.UserCode != "NEW-FIXTURE" || s.snapshot.Auth.ExpiresAt != 456 {
		t.Fatal("current challenge was lost")
	}
}

func TestAuthTransportFailureIsScopedAndCannotRemainPending(t *testing.T) {
	a := AuthState{SessionID: "new", Status: "waiting", UserCode: "FIXTURE"}
	failAuthLogin(&a, "old", errors.New("old failure"))
	if a.Status != "waiting" || a.UserCode != "FIXTURE" {
		t.Fatal("stale failure changed current login")
	}
	failAuthLogin(&a, "new", errors.New("transport unavailable"))
	if a.Status != "failed" || a.Error != "transport unavailable" || a.UserCode != "" {
		t.Fatal("transport failure left stale challenge or pending state")
	}
	a.Status = "succeeded"
	a.Error = ""
	failAuthLogin(&a, "new", errors.New("late failure"))
	if a.Status != "succeeded" || a.Error != "" {
		t.Fatal("late transport failure overwrote success")
	}
}

func TestStartLoginImmediateFailureClearsOldChallenge(t *testing.T) {
	entry := &liveSession{}
	entry.state.snapshot.Auth = AuthState{Support: "supported", Providers: []AuthProvider{{ID: "p", Methods: []AuthMethod{{ID: "browser"}}}}, SessionID: "old", Status: "waiting", URL: "https://example.invalid/old", UserCode: "OLD", VerificationURI: "https://example.invalid/device"}
	m := &Manager{sessions: map[string]*liveSession{"active": entry}, active: "active"}
	if err := m.StartLogin("p", "browser"); err == nil {
		t.Fatal("disconnected login unexpectedly succeeded")
	}
	a := entry.state.snapshot.Auth
	if a.Status != "failed" || a.Error == "" || a.SessionID == "old" || a.URL != "" || a.UserCode != "" || a.VerificationURI != "" || a.ExpiresAt != 0 {
		t.Fatalf("disconnected login left stale pending state: %+v", a)
	}
}
