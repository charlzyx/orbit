package menuproof

// postPair constructs the complete native pair before posting either event.
// It reports posting, not OS processing. The real app state is checked later.
func postPair(construct func(int) uintptr, post func(uintptr), receipt func()) error {
	first, second := construct(0), construct(1)
	if first == 0 || second == 0 {
		return ErrBlocked
	}
	post(first)
	post(second)
	if receipt != nil {
		receipt()
	}
	return nil
}
