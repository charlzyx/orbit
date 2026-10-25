package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestRetryCancelPresentationHidesWithoutChangingRuntimeFacts(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(adaptSnapshot(agent.Snapshot{Connected: true, Retrying: true, RetryDetail: agent.RetryDetail{Active: true, Attempt: 1}, ActiveSession: agent.SessionRow{ID: "one"}, RuntimeID: "runtime-one"}, "/one", nil, nil))
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Cancel retry"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tt.Find("Run status"); ok {
		t.Fatal("source cancel should hide retry presentation immediately")
	}
	if !a.Model.Retrying || !a.Model.RetryDetail.Active || len(actions) != 1 || actions[0].Kind != "AbortRetry" {
		t.Fatal("presentation faked authoritative runtime state", a.Model, actions)
	}
}

func TestRetryCancelPresentationRevisionAndIdentityIsolation(t *testing.T) {
	a := fixture(t)
	a.Options.OnAction = func(Action) {}
	state := agent.Snapshot{Connected: true, Retrying: true, RetryDetail: agent.RetryDetail{Active: true, Attempt: 1, Revision: 1}, ActiveSession: agent.SessionRow{ID: "same-session"}, RuntimeID: "runtime-one"}
	a.SetSnapshot(adaptSnapshot(state, "/one", nil, nil))
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	for _, transition := range []string{"new-event", "new-runtime", "new-session"} {
		if err := tt.Click("Cancel retry"); err != nil {
			t.Fatal(transition, err)
		}
		if _, ok := tt.Find("Cancel retry"); ok {
			t.Fatal("cancel remained", transition)
		}
		// Unchanged snapshots (including late cancellation errors) stay dismissed.
		state.ConnectionError = "late cancellation failure"
		a.SetSnapshot(adaptSnapshot(state, "/one", nil, nil))
		tt.Frame()
		if _, ok := tt.Find("Cancel retry"); ok {
			t.Fatal("unchanged snapshot resurrected cancelled retry")
		}
		switch transition {
		case "new-event":
			state.RetryDetail.Revision++
		case "new-runtime":
			state.RuntimeID = "runtime-two"
		case "new-session":
			state.ActiveSession.ID = "other-session"
		}
		a.SetSnapshot(adaptSnapshot(state, "/one", nil, nil))
		tt.Frame()
		if _, ok := tt.Find("Cancel retry"); !ok {
			t.Fatal("stale dismissal crossed identity/event", transition, tt.Texts())
		}
	}
}

func TestRetryCancelPresentationIgnoresStaleExplicitAction(t *testing.T) {
	a := fixture(t)
	a.Options.OnAction = func(Action) {}
	a.SetSnapshot(adaptSnapshot(agent.Snapshot{Connected: true, Retrying: true, RetryDetail: agent.RetryDetail{Active: true, Attempt: 1, Revision: 1}, ActiveSession: agent.SessionRow{ID: "current"}, RuntimeID: "current-runtime"}, "/one", nil, nil))
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	a.Dispatch(Action{Kind: "AbortRetry", SessionID: "obsolete"})
	tt.Frame()
	if _, ok := tt.Find("Cancel retry"); !ok {
		t.Fatal("stale explicit action hid current session retry")
	}
}
