package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestRunStatusMountedCompaction(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(adaptSnapshot(agent.Snapshot{Connected: true, Compacting: true}, "/fixture", nil, nil))
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if _, ok := tt.Find("Run status"); !ok {
		t.Fatal("compaction strip missing", tt.Texts())
	}
}

func TestRunStatusMountedRetryAndSnapshotSwitch(t *testing.T) {
	a := fixture(t)
	retry := agent.RetryDetail{Active: true, Revision: 1, Attempt: 2, HasMax: true, Max: 5, Error: "rate limited", ErrorProvided: true}
	s := agent.Snapshot{Connected: true, Retrying: true, Compacting: true, RetryDetail: retry, ActiveSession: agent.SessionRow{ID: "one"}}
	a.SetSnapshot(adaptSnapshot(s, "/one", nil, nil))
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	text := "Retrying — attempt 2 of 5 · rate limited"
	if _, ok := tt.Find(text); !ok {
		t.Fatal("observed retry details missing", tt.Texts())
	}
	if _, ok := tt.Find(a.Tr("view.preparing_conversation_context")); ok {
		t.Fatal("retry must take precedence")
	}
	if err := tt.Click("Cancel retry"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "AbortRetry" || !a.Model.RetryDetail.Active {
		t.Fatal("cancel must await observed end", actions, a.Model)
	}
	// A rejected cancellation shows its error, but does not resurrect the
	// cancelled presentation without a new retry event.
	s.ConnectionError = "cancel rejected"
	a.SetSnapshot(adaptSnapshot(s, "/one", nil, nil))
	tt.Frame()
	if _, ok := tt.Find(text); ok {
		t.Fatal("rejection resurrected cancelled presentation")
	}
	// A later attempt replaces the text without stale max from the prior event.
	s.RetryDetail = agent.RetryDetail{Active: true, Revision: 2, Attempt: 3, Error: "unavailable", ErrorProvided: true}
	a.SetSnapshot(adaptSnapshot(s, "/one", nil, nil))
	tt.Frame()
	if _, ok := tt.Find("Retrying — attempt 3 · unavailable"); !ok {
		t.Fatal(tt.Texts())
	}
	// A new active session cannot inherit the parked session's strip.
	a.SetSnapshot(adaptSnapshot(agent.Snapshot{Connected: true, ActiveSession: agent.SessionRow{ID: "two"}}, "/two", nil, nil))
	tt.Frame()
	if _, ok := tt.Find("Run status"); ok {
		t.Fatal("retry leaked into switched session")
	}
	// Restore the parked session, then process the observed completion.
	a.SetSnapshot(adaptSnapshot(s, "/one", nil, nil))
	tt.Frame()
	s.RetryDetail = agent.RetryDetail{}
	s.Retrying = false
	a.SetSnapshot(adaptSnapshot(s, "/one", nil, nil))
	tt.Frame()
	if _, ok := tt.Find("Cancel retry"); ok {
		t.Fatal("late cancel control remained")
	}
	if _, ok := tt.Find(a.Tr("view.preparing_conversation_context")); !ok {
		t.Fatal("compaction did not resume")
	}
	s.Compacting = false
	a.SetSnapshot(adaptSnapshot(s, "/one", nil, nil))
	tt.Frame()
	if _, ok := tt.Find("Run status"); ok {
		t.Fatal("settled strip remained")
	}
}
