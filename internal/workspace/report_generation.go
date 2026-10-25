package workspace

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ReportGenerationPrompt intentionally has no Git or filesystem dependency.
// The active workspace, transcript, diagnostics and attachments must not become
// model context for the app's report form. Only the reporter's notes are input.
func ReportGenerationPrompt(kind, notes string) (GenerationPrompt, Draft, error) {
	noun, rule := "", ""
	switch kind {
	case "bug":
		noun = "GitHub bug report"
		rule = "Describe what the user was doing, what happened, and what was expected."
	case "feature":
		noun = "GitHub feature request"
		rule = "Describe the problem or friction, then the proposed solution."
	case "other":
		noun = "GitHub issue"
		rule = "Describe the issue or request concisely."
	default:
		return GenerationPrompt{}, Draft{}, errors.New("report kind must be bug, feature or other")
	}
	notes = strings.TrimSpace(notes)
	title := ""
	for _, line := range strings.Split(notes, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			runes := []rune(strings.TrimRight(line, ".:;"))
			title = string(runes[:min(72, len(runes))])
			break
		}
	}
	input := boundedText(notes, MaxDraftContextBytes)
	if input == "" {
		input = "(no notes provided)"
	}
	system := "Help the user write a " + noun + ". Turn rough notes into concise, specific prose a maintainer can act on. The title is one line, under about 72 characters, without a trailing period or bug:/feat:/# prefix. " + rule + " Use GitHub-flavored Markdown but no headings: the form supplies them. If the notes include reproduction or verification steps, put them in a numbered list at the end. Do not invent unsupported facts, mention generation, or add commentary. Output exactly one fenced text block: title, a blank line, then body."
	user := fmt.Sprintf("<user-notes>\n%s\n</user-notes>\n\nWrite the %s from only these notes. Return one text code fence: title, blank line, then Markdown body without headings.", input, noun)
	return GenerationPrompt{System: system, User: user}, Draft{Title: title, Body: notes}, nil
}

// GenerateReportDraftRPC mirrors the source notes-only report generator while
// using the user's requested isolated RPC transport rather than pi -p. It never
// submits a report. A model failure returns an explicitly marked editable draft.
func GenerateReportDraftRPC(ctx context.Context, cfg GenerationConfig, kind, notes string) (Draft, error) {
	prompt, fallback, err := ReportGenerationPrompt(kind, notes)
	if err != nil {
		return Draft{}, err
	}
	raw, err := isolatedRPC(ctx, cfg, prompt)
	if err == nil {
		var result Draft
		result, err = ParseDraft(raw)
		if err == nil {
			return result, nil
		}
	}
	if ctx.Err() != nil {
		return Draft{}, ctx.Err()
	}
	fallback.Fallback, fallback.Warning = true, err.Error()
	return fallback, nil
}
