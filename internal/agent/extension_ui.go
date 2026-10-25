package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

// ApprovalRequest is the guard extension's marked select, not a modal dialog.
// The extension owns persistent, per-access-mode allowlisting; the host sends
// its exact option string rather than inventing a session permission.
type ApprovalRequest struct {
	ID, Tool, Detail string
	Options          []string
}
type ExtensionWidget struct {
	Key, Placement string
	Lines          []string
}
type ExtensionNotice struct {
	ID       string
	Sequence uint64
	Message  string
}
type AskOption struct{ Label, Description, Preview string }
type AskQuestion struct {
	Header, Question string
	MultiSelect      bool
	Options          []AskOption
}
type AskAnswer struct {
	Kind        string
	Option      int
	Checked     []bool
	Custom      string
	CustomLabel string
}
type AskPrompt struct {
	ID, ToolCallID string
	Questions      []AskQuestion
	Submitted      bool
}
type askReplay struct {
	questions     []AskQuestion
	answers       []AskAnswer
	index         int
	pendingCustom *string
	currentID     string
}

func IsAskTool(name string) bool {
	return name == "ask_user_question" || name == "ask_question" || name == "ask_user"
}

// ParseAskQuestions preserves the reference's tolerant per-entry parsing.
func ParseAskQuestions(raw json.RawMessage) []AskQuestion {
	var args map[string]json.RawMessage
	if json.Unmarshal(raw, &args) != nil {
		return nil
	}
	var list []json.RawMessage
	if json.Unmarshal(args["questions"], &list) != nil {
		return nil
	}
	text := func(v map[string]json.RawMessage, key string) string {
		var s string
		_ = json.Unmarshal(v[key], &s)
		return s
	}
	var out []AskQuestion
	for _, raw := range list {
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil {
			continue
		}
		var question *string
		if json.Unmarshal(fields["question"], &question) != nil || question == nil {
			continue
		}
		q := AskQuestion{Question: *question, Header: text(fields, "header")}
		_ = json.Unmarshal(fields["multiSelect"], &q.MultiSelect)
		var options []json.RawMessage
		_ = json.Unmarshal(fields["options"], &options)
		for _, rawOption := range options {
			var opt map[string]json.RawMessage
			if json.Unmarshal(rawOption, &opt) != nil {
				continue
			}
			var label *string
			if json.Unmarshal(opt["label"], &label) != nil || label == nil {
				continue
			}
			q.Options = append(q.Options, AskOption{*label, text(opt, "description"), text(opt, "preview")})
		}
		out = append(out, q)
	}
	return out
}

func (s *sessionState) askToolEvent(event pirpc.Event) {
	id := event.String("toolCallId")
	if event.Type == "tool_execution_start" && IsAskTool(event.String("toolName")) {
		s.askToolID = id
		s.askQuestions = ParseAskQuestions(event.Fields["args"])
		s.askReplay = nil
	} else if event.Type == "tool_execution_end" && id != "" && id == s.askToolID {
		s.askToolID = ""
		s.askQuestions = nil
		s.askReplay = nil
		s.snapshot.Ask = nil
	}
}
func (s *sessionState) blockingDialog() bool {
	if s.snapshot.Approval != nil {
		return true
	}
	for _, d := range s.snapshot.Dialogs {
		if d.Method != "custom" {
			return true
		}
	}
	return false
}
func (s *sessionState) cancelUI(id string) {
	if id != "" {
		s.uiResponses = append(s.uiResponses, pirpc.UIResponse{ID: id, Cancelled: true})
	}
}
func (s *sessionState) routeSpecialDialog(event pirpc.Event) bool {
	method, id := event.String("method"), event.ID
	if s.askReplay != nil && (method == "select" || method == "input") {
		s.replayAsk(id, method)
		return true
	}
	if method == "select" && strings.HasPrefix(event.String("title"), "[orbit-guard] ") {
		if s.blockingDialog() {
			s.cancelUI(id)
			return true
		}
		rest := strings.TrimPrefix(event.String("title"), "[orbit-guard] ")
		tool, detail, ok := strings.Cut(rest, "\t")
		if !ok {
			tool, detail = "", rest
		}
		options := extensionStringOptions(event.Fields["options"])
		s.snapshot.Approval = &ApprovalRequest{ID: id, Tool: tool, Detail: detail, Options: options}
		return true
	}
	if s.askToolID != "" && len(s.askQuestions) > 0 && (method == "select" || method == "input") {
		if s.blockingDialog() || s.snapshot.Ask != nil {
			s.cancelUI(id)
			return true
		}
		s.snapshot.Ask = &AskPrompt{ID: id, ToolCallID: s.askToolID, Questions: s.askQuestions}
		return true
	}
	return false
}
func (r *askReplay) reply(method string) (string, bool) {
	if r.index >= len(r.questions) || r.index >= len(r.answers) {
		return "", false
	}
	q, a := r.questions[r.index], r.answers[r.index]
	if method == "select" {
		switch a.Kind {
		case "option":
			if a.Option < 0 || a.Option >= len(q.Options) {
				return "", false
			}
			r.index++
			return fmt.Sprintf("%d. %s", a.Option+1, q.Options[a.Option].Label), true
		case "custom":
			text := a.Custom
			r.pendingCustom = &text
			label := a.CustomLabel
			if label == "" {
				label = "Type something."
			}
			return fmt.Sprintf("%d. %s", len(q.Options)+1, label), true
		}
	} else if r.pendingCustom != nil {
		text := *r.pendingCustom
		r.pendingCustom = nil
		r.index++
		return text, true
	} else if a.Kind == "multi" {
		r.index++
		if custom := strings.TrimSpace(a.Custom); custom != "" {
			return custom, true
		}
		var indices []string
		for i, on := range a.Checked {
			if on {
				indices = append(indices, fmt.Sprint(i+1))
			}
		}
		return strings.Join(indices, ","), true
	}
	return "", false
}
func (s *sessionState) replayAsk(id, method string) {
	r := s.askReplay
	if r == nil {
		return
	}
	r.currentID = id
	value, ok := r.reply(method)
	if !ok {
		s.cancelUI(id)
		s.askReplay = nil
		return
	}
	s.uiResponses = append(s.uiResponses, pirpc.UIResponse{ID: id, Value: &value})
	if r.index >= len(r.questions) {
		s.askReplay = nil
	}
}
func (s *sessionState) submitAsk(id string, answers []AskAnswer) error {
	p := s.snapshot.Ask
	if p == nil || p.ID != id || p.Submitted {
		return fmt.Errorf("questionnaire is no longer pending")
	}
	if len(answers) != len(p.Questions) {
		return fmt.Errorf("every question needs an answer")
	}
	for i, q := range p.Questions {
		a := answers[i]
		if q.MultiSelect {
			if a.Kind != "multi" || len(a.Checked) != len(q.Options) {
				return fmt.Errorf("invalid multi-select answer")
			}
		} else if a.Kind != "custom" && (a.Kind != "option" || a.Option < 0 || a.Option >= len(q.Options)) {
			return fmt.Errorf("invalid single-select answer")
		}
	}
	// Take a copy: a UI may reuse its answer buffers before the next request.
	copyAnswers := make([]AskAnswer, len(answers))
	copy(copyAnswers, answers)
	for i := range copyAnswers {
		copyAnswers[i].Checked = append([]bool(nil), answers[i].Checked...)
	}
	p.Submitted = true
	s.askReplay = &askReplay{questions: p.Questions, answers: copyAnswers, currentID: id}
	method := "select"
	if p.Questions[0].MultiSelect {
		method = "input"
	}
	s.replayAsk(id, method)
	return nil
}
func (s *sessionState) applyWidget(event pirpc.Event) {
	var keyValue *string
	if event.DecodeField("widgetKey", &keyValue) != nil || keyValue == nil {
		return
	}
	key := *keyValue
	placement := "aboveEditor"
	if event.String("widgetPlacement") == "belowEditor" {
		placement = "belowEditor"
	}
	var raw []json.RawMessage
	_ = event.DecodeField("widgetLines", &raw)
	lines := make([]string, len(raw))
	for i, value := range raw {
		_ = json.Unmarshal(value, &lines[i])
	}
	for i, w := range s.snapshot.ExtensionWidgets {
		if w.Key != key {
			continue
		}
		if len(lines) == 0 {
			s.snapshot.ExtensionWidgets = append(s.snapshot.ExtensionWidgets[:i], s.snapshot.ExtensionWidgets[i+1:]...)
		} else {
			s.snapshot.ExtensionWidgets[i] = ExtensionWidget{key, placement, lines}
		}
		return
	}
	if len(lines) > 0 {
		s.snapshot.ExtensionWidgets = append(s.snapshot.ExtensionWidgets, ExtensionWidget{key, placement, lines})
	}
}

// cancelExtensionUI is called before parking a session, not just after its next
// request. The previous process must not stay blocked on a hidden surface.
// Widgets belong to the parked process and must survive a warm resume.
func (s *sessionState) cancelExtensionUI() {
	for _, d := range s.snapshot.Dialogs {
		s.cancelUI(d.ID)
	}
	if p := s.snapshot.Approval; p != nil {
		s.cancelUI(p.ID)
	}
	if p := s.snapshot.Ask; p != nil && !p.Submitted {
		s.cancelUI(p.ID)
	}
	widgets := s.snapshot.ExtensionWidgets
	s.clearExtensionUI()
	s.snapshot.ExtensionWidgets = widgets
}
func (s *sessionState) clearExtensionUI() {
	s.snapshot.Dialogs = nil
	s.snapshot.Approval = nil
	s.snapshot.Ask = nil
	s.snapshot.ExtensionWidgets = nil
	s.snapshot.ExtensionNotices = nil
	s.snapshot.ExtensionStatusText = ""
	s.snapshot.ExtensionStatusAt = time.Time{}
	s.askToolID = ""
	s.askQuestions = nil
	s.askReplay = nil
}
func (s *sessionState) takeUIResponses() []pirpc.UIResponse {
	responses := s.uiResponses
	s.uiResponses = nil
	return responses
}

// SubmitAsk commits the locally buffered questionnaire, then subsequent Pi
// primitives replay from it. This is not a prompt or an invented completed tool.
func (m *Manager) SubmitAsk(id string, answers []AskAnswer) error {
	m.mu.Lock()
	entry := m.sessions[m.active]
	if entry == nil || entry.client == nil {
		m.mu.Unlock()
		return fmt.Errorf("Pi session is not ready")
	}
	err := entry.state.submitAsk(id, answers)
	m.mu.Unlock()
	if err != nil {
		return err
	}
	m.Tick()
	m.changed()
	return nil
}

func (s *sessionState) setComposerText(text string) {
	s.snapshot.ComposerText = text
	s.snapshot.ComposerRevision++
}

func extensionStringOptions(raw json.RawMessage) []string {
	var values []json.RawMessage
	_ = json.Unmarshal(raw, &values)
	var out []string
	for _, v := range values {
		var text *string
		if json.Unmarshal(v, &text) == nil && text != nil {
			out = append(out, *text)
		}
	}
	return out
}
