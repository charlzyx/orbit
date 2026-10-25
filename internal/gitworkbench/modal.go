package gitworkbench

// HasModal reports the real native dialog/confirmation state to host menu
// dispatch. A click outside the tree (for example an OS menu) must not bypass
// the same modal guard as an in-window shortcut.
func (s *State) HasModal() bool { return s != nil && (s.dialogOpen || s.confirmOpen) }
