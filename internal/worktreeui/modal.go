package worktreeui

// HasModal lets native host menu dispatch honor this page's active dialog.
func (s *State) HasModal() bool { return s != nil && (s.dialogOpen || s.confirmOpen) }
