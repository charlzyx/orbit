package worktreeui

import "strings"

// text is UI-thread only: the host may resolve it from its active locale.
func (s *State) text(key, fallback string, values ...string) string {
	value := fallback
	if s.Options.Translate != nil {
		if translated := s.Options.Translate(key); translated != "" && translated != key {
			value = translated
		}
	}
	for i := 0; i+1 < len(values); i += 2 {
		value = strings.ReplaceAll(value, "%{"+values[i]+"}", values[i+1])
	}
	return value
}
