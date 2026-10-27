package bugreportui

import "strings"

// Display-only local messages: operation identity and arbitrary backend details stay intact.
func (s *State) displayMessage(message string) string {
	// Status is constructed locally; translate only its owned prefix, retaining warning/URL/path bytes.
	const draft = "Local draft used. Review and edit it before sharing."
	if strings.HasPrefix(message, draft+" ") {
		return s.label("go.bug.status_4", draft) + strings.TrimPrefix(message, draft)
	}
	if strings.HasPrefix(message, "Issue created: ") || strings.HasPrefix(message, "GitHub draft prepared. Review and submit it in your browser. Screenshots: ") {
		before, path, hasPath := strings.Cut(message, " Screenshots: ")
		if strings.HasPrefix(before, "Issue created: ") {
			before = s.label("go.bug.issue_created", "Issue created:") + " " + strings.TrimPrefix(before, "Issue created: ")
		} else {
			before = s.label("go.bug.status_15", before)
		}
		if hasPath {
			before += " " + s.label("go.bug.screenshots", "Screenshots:") + " " + path
		}
		return before
	}
	switch message {
	case "Cancelling":
		return s.label("go.bug.cancelling", message)
	case "Cancelled. Your report has been kept.":
		return s.label("go.bug.status_0", message)
	case "Submission stopped. GitHub may already have created the issue; check the repository before retrying.":
		return s.label("go.bug.status_1", message)
	case "Collecting diagnostics":
		return s.label("go.bug.status_2", message)
	case "Drafting from your notes":
		return s.label("go.bug.status_3", message)
	case "Local draft used. Review and edit it before sharing.":
		return s.label("go.bug.status_4", message)
	case "Up to 6 screenshots.":
		return s.label("go.bug.status_5", message)
	case "Reading selected screenshots":
		return s.label("go.bug.status_6", message)
	case "Choose screenshots":
		return s.label("go.bug.status_7", message)
	case "Capturing current Orbit window":
		return s.label("go.bug.status_8", message)
	case "use x,y,width,height; separate boxes with semicolons":
		return s.label("go.bug.status_9", message)
	case "redaction coordinates must be nonnegative pixels":
		return s.label("go.bug.status_10", message)
	case "redaction width and height must be positive":
		return s.label("go.bug.status_11", message)
	case "Redacting screenshot":
		return s.label("go.bug.status_12", message)
	case "Preparing local preview":
		return s.label("go.bug.status_13", message)
	case "Submitting reviewed report":
		return s.label("go.bug.status_14", message)
	case "GitHub draft prepared. Review and submit it in your browser.":
		return s.label("go.bug.status_15", message)
	default:
		return message
	}
}

// Only validation created by this package is eligible for local display translation.
type localValidationError string

func (e localValidationError) Error() string { return string(e) }
func (s *State) errorMessage(err error) string {
	if e, ok := err.(localValidationError); ok {
		return s.displayMessage(string(e))
	}
	return err.Error()
}
