package gitworkbench

// Rust owns independent Issue and PR ComposerInput entities. The native form
// shares widgets, so keep the same independent drafts when cancelling, closing,
// or visiting a comment form. These are session-local, never credentials.
type githubFormDraft struct {
	title, body, notes, labels, assignees, base string
	draft                                       bool
	attachments                                 []string
}

func (s *State) githubDraft(kind string) *githubFormDraft {
	if kind == "new-issue" {
		return &s.issueDraft
	}
	if kind == "new-pull" {
		return &s.pullDraft
	}
	return nil
}
func (s *State) rememberGitHubDraft() {
	if d := s.githubDraft(s.dialog); d != nil {
		*d = githubFormDraft{title: s.title, body: s.body, notes: s.notes, labels: s.labels, assignees: s.assignees, base: s.base, draft: s.draft, attachments: append([]string(nil), s.attachments...)}
	}
}
func (s *State) restoreGitHubDraft(kind string) {
	if d := s.githubDraft(kind); d != nil {
		s.title, s.body, s.notes, s.labels, s.assignees, s.base = d.title, d.body, d.notes, d.labels, d.assignees, d.base
		s.draft = d.draft
		s.attachments = append([]string(nil), d.attachments...)
	}
}
func (s *State) finishGitHubDraft(kind string) {
	if d := s.githubDraft(kind); d != nil {
		d.title, d.body, d.notes, d.labels, d.assignees = "", "", "", "", ""
		d.attachments = nil
	}
	if s.dialog == kind {
		s.dialogOpen = false
		s.title, s.body, s.notes, s.labels, s.assignees = "", "", "", "", ""
		s.attachments = nil
	}
}
