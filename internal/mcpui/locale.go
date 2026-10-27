package mcpui

import "strings"

// Translate only locally-owned validation, preserving raw external errors.
func (s *State) displayValidation(message string) string {
	switch message {
	case "choose global or project scope":
		return s.label("go.mcpui.validation_0", message)
	case "select a workspace before saving project configuration":
		return s.label("go.mcpui.validation_1", message)
	case "timeout must be a positive integer":
		return s.label("go.mcpui.validation_2", message)
	case "OAuth callback port must be 1\u201365535":
		return s.label("go.mcpui.validation_3", message)
	case "OAuth callback URL must use loopback HTTP":
		return s.label("go.mcpui.validation_4", message)
	case "choose stdio or streamable HTTP transport":
		return s.label("go.mcpui.validation_5", message)
	case "server name must contain only letters, digits, hyphen or underscore":
		return s.label("go.mcpui.local_0", message)
	case "server definition must be an object":
		return s.label("go.mcpui.local_1", message)
	case "type must be a string":
		return s.label("go.mcpui.local_2", message)
	case "legacy SSE is unsupported":
		return s.label("go.mcpui.local_3", message)
	case "enabled must be boolean":
		return s.label("go.mcpui.local_4", message)
	case "invalid exposure mode":
		return s.label("go.mcpui.local_5", message)
	case "toolExposure must be an object":
		return s.label("go.mcpui.local_6", message)
	case "invalid tool exposure mode":
		return s.label("go.mcpui.local_7", message)
	case "server URL must be absolute http or https":
		return s.label("go.mcpui.local_8", message)
	case "transport type does not match URL":
		return s.label("go.mcpui.local_9", message)
	case "oauth must be an object":
		return s.label("go.mcpui.local_10", message)
	case "OAuth callback must use loopback HTTP":
		return s.label("go.mcpui.local_11", message)
	case "a valid command is required":
		return s.label("go.mcpui.local_12", message)
	case "transport type does not match command":
		return s.label("go.mcpui.local_13", message)
	case "args must be an array":
		return s.label("go.mcpui.local_14", message)
	case "args must contain strings":
		return s.label("go.mcpui.local_15", message)
	case "cwd must be a valid string":
		return s.label("go.mcpui.local_16", message)
	default:
		return strings.ReplaceAll(s.label("go.mcp.validation_detail", "Invalid configuration: %{detail}"), "%{detail}", message)
	}
}

func (s *State) displayActionError() string {
	if s.localBusyError {
		return s.label("go.mcp.busy", s.Error)
	}
	return s.Error
}
