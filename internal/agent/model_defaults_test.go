package agent

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/charlzyx/orbit/internal/pirpc"
	"github.com/charlzyx/orbit/internal/preferences"
)

func birthFixture() (birthDefault, Snapshot) {
	return birthDefault{armed: true, loaded: true, value: preferences.SessionDefault{Provider: "p", ModelID: "target", Thinking: "high"}}, Snapshot{ActiveSession: SessionRow{ID: "new"}, Model: ModelChoice{Provider: "p", ID: "old"}, ThinkingLevel: "off"}
}
func TestBirthDefaultWaitsForIdentityCatalogAndLiveModel(t *testing.T) {
	d, s := birthFixture()
	s.ActiveSession.ID = ""
	if d.next(s) != nil || !d.armed {
		t.Fatal("before identity")
	}
	s.ActiveSession.ID = "new"
	if d.next(s) != nil || !d.armed {
		t.Fatal("empty catalog must stay armed")
	}
	s.Models = []ModelChoice{{Provider: "p", ID: "target", Reasoning: true}}
	c := d.next(s)
	if c == nil || c.Type != pirpc.SetModel {
		t.Fatal(c)
	}
	if d.next(s) != nil {
		t.Fatal("duplicate set_model")
	}
	s.Model = s.Models[0]
	c = d.next(s)
	if c == nil || c.Type != pirpc.SetThinkingLevel || d.armed {
		t.Fatal(c, d)
	}
	if d.next(s) != nil {
		t.Fatal("duplicate thinking")
	}
}
func TestBirthDefaultSkipsRemovedUnsupportedAlreadyLiveAndUnset(t *testing.T) {
	t.Run("removed", func(t *testing.T) {
		d, s := birthFixture()
		s.Models = []ModelChoice{{Provider: "other", ID: "target"}}
		if d.next(s) != nil || d.armed {
			t.Fatal(d)
		}
	})
	t.Run("unsupported", func(t *testing.T) {
		d, s := birthFixture()
		s.Models = []ModelChoice{{Provider: "p", ID: "target", Reasoning: true, Raw: json.RawMessage(`{"thinkingLevelMap":{"high":null}}`)}}
		s.Model = s.Models[0]
		if d.next(s) != nil || d.armed {
			t.Fatal(d)
		}
	})
	t.Run("already live", func(t *testing.T) {
		d, s := birthFixture()
		s.Models = []ModelChoice{{Provider: "p", ID: "target", Reasoning: true}}
		s.Model = s.Models[0]
		s.ThinkingLevel = "high"
		if d.next(s) != nil || d.armed {
			t.Fatal(d)
		}
	})
	t.Run("unset", func(t *testing.T) {
		d, s := birthFixture()
		d.value.ModelID = ""
		if d.next(s) != nil || d.armed {
			t.Fatal(d)
		}
	})
}
func TestBirthDefaultProviderlessSelectsFirstSourceCatalogMatch(t *testing.T) {
	d, s := birthFixture()
	d.value.Provider = ""
	s.Models = []ModelChoice{{Provider: "first", ID: "target"}, {Provider: "second", ID: "target"}}
	c := d.next(s)
	if c == nil || c.Params["provider"] != "first" {
		t.Fatal(c)
	}
}
func TestCatalogThinkingLevelsSourceCapabilitySemantics(t *testing.T) {
	if got := CatalogThinkingLevels(ModelChoice{}); !slices.Equal(got, []string{"off"}) {
		t.Fatal(got)
	}
	if got := CatalogThinkingLevels(ModelChoice{Reasoning: true}); !slices.Equal(got, []string{"off", "minimal", "low", "medium", "high"}) {
		t.Fatal(got)
	}
	got := CatalogThinkingLevels(ModelChoice{Reasoning: true, Raw: json.RawMessage(`{"thinkingLevelMap":{"off":null,"minimal":null,"medium":null,"xhigh":null,"max":"highest"}}`)})
	if !slices.Equal(got, []string{"low", "high", "max"}) {
		t.Fatal(got)
	}
}
func TestManualModelIntentDisarmsPendingDefaultWhileNotReady(t *testing.T) {
	for _, kind := range []string{pirpc.SetModel, pirpc.SetThinkingLevel, pirpc.CycleModel, pirpc.CycleThinkingLevel} {
		t.Run(kind, func(t *testing.T) {
			m := testManager(t, nil)
			m.mu.Lock()
			e := &liveSession{loading: true, modelDefault: birthDefault{armed: true}}
			m.sessions["draft"] = e
			m.active = "draft"
			m.mu.Unlock()
			if err := m.Send(kind, nil); err == nil {
				t.Fatal("not-ready accepted")
			}
			m.mu.Lock()
			defer m.mu.Unlock()
			if e.modelDefault.armed || e.modelGeneration != 1 {
				t.Fatal(e.modelDefault, e.modelGeneration)
			}
		})
	}
}
