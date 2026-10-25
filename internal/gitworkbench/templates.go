package gitworkbench

import (
	"context"
	"strings"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func (s *State) formTemplates(pull bool) []workspace.RepositoryTemplate {
	if pull {
		return s.Data.PullTemplates
	}
	return s.Data.IssueTemplates
}
func (s *State) applyTemplate(pull bool, index int) {
	templates := s.formTemplates(pull)
	if index < 0 || index >= len(templates) {
		if s.body == s.templateOriginal {
			s.body = ""
		}
		s.templateIndex = -1
		s.templateOriginal = ""
		return
	}
	template := templates[index]
	s.templateIndex = index
	s.templateOriginal = template.Body
	s.body = template.Body
	if strings.TrimSpace(s.title) == "" {
		s.title = template.TitlePrefix
	}
	s.labels = strings.Join(template.Labels, ", ")
}
func (s *State) applyLateTemplate() {
	if !s.dialogOpen || (s.dialog != "new-issue" && s.dialog != "new-pull") || s.templateIndex >= 0 || s.body != "" {
		return
	}
	pull := s.dialog == "new-pull"
	if len(s.formTemplates(pull)) == 1 {
		s.applyTemplate(pull, 0)
	}
}
func (s *State) templatePicker(c *ui.Context, pull bool) {
	templates := s.formTemplates(pull)
	if len(templates) == 0 {
		return
	}
	label := "No template"
	if s.templateIndex >= 0 && s.templateIndex < len(templates) {
		label = templates[s.templateIndex].Label()
	}
	button(c, label+" ▾", s.Busy).Menu(func(m *ui.Menu) {
		if m.Item("No template").Chosen() {
			s.applyTemplate(pull, -1)
		}
		for i, template := range templates {
			if m.Item(template.Label()).Chosen() {
				s.applyTemplate(pull, i)
			}
		}
	})
}
func (s *State) generateDraft(pull bool) {
	request := workspace.DraftRequest{Kind: "issue", Notes: s.notes, Base: s.base}
	if pull {
		request.Kind = "pull"
	}
	templates := s.formTemplates(pull)
	if s.templateIndex >= 0 && s.templateIndex < len(templates) {
		template := templates[s.templateIndex]
		request.Template = &template
	}
	generate := s.Options.GenerateDraft
	s.start("Generating draft", func(ctx context.Context, _ workspace.Git, _ workspace.GitHub) (func(*State), string, error) {
		draft, e := generate(ctx, request)
		return func(s *State) {
			if e == nil {
				s.title = draft.Title
				s.body = draft.Body
				if draft.Fallback {
					s.Status = "Editable fallback: " + draft.Warning
				}
			}
		}, "", e
	})
}
