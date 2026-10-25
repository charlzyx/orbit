package agent

import (
	"bytes"
	"encoding/json"
	"slices"

	"github.com/charlzyx/orbit/internal/pirpc"
	"github.com/charlzyx/orbit/internal/preferences"
)

// CatalogThinkingLevels follows the pinned original catalog capability map.
// A null removes a level; xhigh/max require an explicit non-null mapping.
func CatalogThinkingLevels(model ModelChoice) []string {
	if !model.Reasoning {
		return []string{"off"}
	}
	var raw struct {
		Map map[string]json.RawMessage `json:"thinkingLevelMap"`
	}
	_ = json.Unmarshal(model.Raw, &raw)
	var levels []string
	for _, level := range []string{"off", "minimal", "low", "medium", "high", "xhigh", "max"} {
		value, ok := raw.Map[level]
		if ok && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			continue
		}
		if !ok && (level == "xhigh" || level == "max") {
			continue
		}
		levels = append(levels, level)
	}
	return levels
}

type birthDefault struct {
	armed          bool
	loaded         bool
	value          preferences.SessionDefault
	session        string
	pushedModel    preferences.Model
	pushedThinking string
}

// next is I/O-free and called only after authoritative state/catalog responses.
// It deliberately waits for the selected model to become live before thinking:
// Pi resets the thinking level while applying a model switch.
func (d *birthDefault) next(s Snapshot) *pirpc.Command {
	if !d.armed || !d.loaded || s.ActiveSession.ID == "" {
		return nil
	}
	if d.value.ModelID == "" {
		d.armed = false
		return nil
	}
	if d.session != s.ActiveSession.ID {
		d.session = s.ActiveSession.ID
		d.pushedModel = preferences.Model{}
		d.pushedThinking = ""
	}
	var model *ModelChoice
	for i := range s.Models {
		m := &s.Models[i]
		if m.ID == d.value.ModelID && (d.value.Provider == "" || m.Provider == d.value.Provider) {
			model = m
			break
		}
	}
	if model == nil {
		if len(s.Models) > 0 {
			d.armed = false
		}
		return nil
	}
	wanted := preferences.Model{Provider: model.Provider, ID: model.ID}
	if s.Model.Provider != model.Provider || s.Model.ID != model.ID {
		if d.pushedModel == wanted {
			return nil
		}
		d.pushedModel = wanted
		command := pirpc.NewCommand(pirpc.SetModel, map[string]any{"provider": model.Provider, "modelId": model.ID})
		return &command
	}
	d.armed = false
	if d.value.Thinking != "" && s.ThinkingLevel != d.value.Thinking && d.pushedThinking != d.value.Thinking && slices.Contains(CatalogThinkingLevels(*model), d.value.Thinking) {
		d.pushedThinking = d.value.Thinking
		command := pirpc.NewCommand(pirpc.SetThinkingLevel, map[string]any{"level": d.value.Thinking})
		return &command
	}
	return nil
}
