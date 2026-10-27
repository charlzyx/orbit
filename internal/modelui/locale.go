package modelui

// Translate only locally-owned validation, preserving raw external errors.
func (s *State) displayValidation(message string) string {
	switch message {
	case "Model catalog update is not connected.":
		return s.label("go.modelui.validation_0", message)
	case "Pi reported an incomplete model identity.":
		return s.label("go.modelui.validation_1", message)
	case "Model selection is not connected.":
		return s.label("go.modelui.validation_2", message)
	case "Thinking selection is not connected.":
		return s.label("go.modelui.validation_3", message)
	default:
		return message
	}
}
