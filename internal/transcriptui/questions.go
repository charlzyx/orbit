package transcriptui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
)

type QuestionOption struct{ Label, Description, Preview string }
type Question struct {
	Header, Text string
	Multi        bool
	Options      []QuestionOption
}

func (t ToolCall) Questions() []Question {
	if !in(t.Name, "ask_user_question ask_question ask_user") {
		return nil
	}
	var list []object
	_ = json.Unmarshal(obj(t.Arguments)["questions"], &list)
	var qs []Question
	for _, q := range list {
		var text string
		if json.Unmarshal(q["question"], &text) != nil {
			continue
		}
		v := Question{Header: str(q["header"]), Text: text, Multi: boolean(q["multiSelect"])}
		var options []object
		_ = json.Unmarshal(q["options"], &options)
		for _, o := range options {
			var label string
			if json.Unmarshal(o["label"], &label) == nil {
				v.Options = append(v.Options, QuestionOption{label, str(o["description"]), str(o["preview"])})
			}
		}
		qs = append(qs, v)
	}
	return qs
}
func (t ToolCall) QuestionAnswer(question string) string {
	envelope := str(t.Output)
	if !strings.HasPrefix(envelope, "User has answered your questions:") {
		return ""
	}
	needle := "\"" + question + "\"=\""
	i := strings.Index(envelope, needle)
	if i < 0 {
		return ""
	}
	rest := envelope[i+len(needle):]
	end := strings.Index(rest, "\". ")
	if end < 0 {
		end = strings.Index(rest, "\".")
	}
	if end < 0 {
		end = len(rest)
	}
	return strings.TrimSpace(rest[:end])
}
func answerIncludes(answer, label string) bool {
	for _, part := range strings.Split(answer, ",") {
		if strings.TrimSpace(part) == label {
			return true
		}
	}
	return false
}
func optionAnswer(q Question, answer string) bool {
	any := false
	for _, part := range strings.Split(answer, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		any = true
		found := false
		for _, o := range q.Options {
			found = found || o.Label == part
		}
		if !found {
			return false
		}
	}
	return any
}

// questionCard is deliberately a read-only record. Live authorization answers
// must continue through the host's actual extension_ui_request controller.
func questionCard(c *ui.Context, t ToolCall, questions []Question, o Options) ui.Element {
	declined := strings.Contains(str(t.Output), "declined to answer")
	waiting := !present(t.Output) && !t.Failed
	return ui.Column(c.Key("question:"+t.ID)).AlignSelf(ui.Stretch).MinWidth(0).Border(1, o.Colors.Border).Radius(8).Background(o.Colors.Surface).Children(func() {
		ui.Row(c).AlignSelf(ui.Stretch).Unselectable().Height(32).Padding(0, 8).Gap(8).AlignItems(ui.Center).Children(func() {
			o.icon(c, "icons/task.svg", 14, o.Colors.Muted)
			ui.Text(c, o.tr("transcript_view.question", "Question")).FontWeight(600).TextColor(o.Colors.Accent)
			if len(questions) > 1 {
				ui.Text(c, o.tr("transcript.asked", "%{count} asked", "count", fmt.Sprint(len(questions)))).FontSize(12).TextColor(o.Colors.Muted)
			}
			ui.Spacer(c)
			if declined {
				ui.Text(c, o.tr("transcript_view.no_answer", "No answer")).TextColor(o.Colors.Muted).FontSize(12)
			} else if t.Failed {
				o.icon(c, "icons/stop.svg", 12, o.Colors.Error)
			}
		})
		ui.Column(c).AlignSelf(ui.Stretch).Padding(0, 10, 10).Gap(10).Children(func() {
			for i, q := range questions {
				ui.Column(c.Key(i)).AlignSelf(ui.Stretch).Gap(4).Children(func() {
					ui.Row(c).AlignSelf(ui.Stretch).Gap(6).AlignItems(ui.Center).Children(func() {
						if strings.TrimSpace(q.Header) != "" {
							ui.Text(c, q.Header).FontSize(12).Padding(1, 6).Radius(4).Background(o.Colors.Border).TextColor(o.Colors.Muted)
						}
						ui.Text(c, q.Text).Selectable().Grow(1).MinWidth(0).FontWeight(500).TextColor(o.Colors.Text)
					})
					answer := t.QuestionAnswer(q.Text)
					for j, option := range q.Options {
						chosen := answer != "" && answerIncludes(answer, option.Label)
						row := ui.Row(c.Key(j)).AlignSelf(ui.Stretch).Gap(8).Padding(4, 8).Radius(6)
						if chosen {
							row.Background(o.Colors.Selection)
						}
						row.Children(func() {
							ui.Text(c, fmt.Sprintf("%d.", j+1)).Unselectable().FontSize(12).TextColor(o.Colors.Muted)
							ui.Text(c, option.Label).Selectable().Grow(1).MinWidth(0).TextColor(o.Colors.Text)
							if chosen {
								o.icon(c, "icons/check.svg", 12, o.Colors.Accent)
							}
						})
					}
					if answer != "" && !optionAnswer(q, answer) {
						ui.Row(c).AlignSelf(ui.Stretch).Padding(4, 8).Gap(8).Background(o.Colors.Accent.Alpha(.1)).Radius(6).Children(func() {
							ui.Text(c, "A.").Unselectable().FontSize(12).TextColor(o.Colors.Accent)
							ui.Text(c, answer).Selectable().Grow(1).MinWidth(0).TextColor(o.Colors.Text)
						})
					}
					if waiting {
						ui.Text(c, o.tr("transcript_view.waiting_for_an_answer", "Waiting for an answer…")).Unselectable().FontSize(12).TextColor(o.Colors.Muted).Padding(0, 8)
					}
				})
			}
		})
	})
}
