package workspace

// Generation deliberately uses a fresh, tool-free Pi RPC process, not pi -p
// and not the active chat session. Only the editable result is returned; these
// helpers never commit, publish an issue, push, or modify the Git index.
import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/charlzyx/orbit/internal/pirpc"
)

const (
	MaxGenerationDiffBytes          = 96 * 1024
	MaxGenerationOriginalBytes      = 48 * 1024
	MaxGenerationTotalOriginalBytes = 192 * 1024
	MaxDraftContextBytes            = 24 * 1024
)

type GenerationConfig struct {
	Pi              pirpc.Config
	Provider, Model string
	Timeout         time.Duration
	ContextWindow   int // Commit only; zero means unknown, not the active model window.
}
type GeneratedText struct {
	Text     string
	Fallback bool
	Warning  string
}
type Draft struct {
	Title, Body string
	Fallback    bool
	Warning     string
}
type DraftRequest struct {
	Kind, Notes, Base string
	Template          *RepositoryTemplate
}
type GenerationPrompt struct{ System, User string }

func boundedText(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 0 {
		max = 0
	}
	for max > 0 && !utf8.RuneStart(s[max]) {
		max--
	}
	return s[:max] + "\n… [truncated]"
}
func isolatedRPC(ctx context.Context, cfg GenerationConfig, prompt GenerationPrompt) (string, error) {
	if e := ctx.Err(); e != nil {
		return "", e
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 180 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	config := cfg.Pi
	config.NoSession = true
	config.SessionDir = ""
	config.TrustProject = false
	config.Args = append(append([]string(nil), config.Args...), "--no-tools", "--no-skills", "--no-context-files", "--system-prompt", prompt.System)
	env := map[string]string{}
	for k, v := range config.EnvOverrides {
		env[k] = v
	}
	env["PI_SKIP_VERSION_CHECK"] = "1"
	env["NO_COLOR"] = "1"
	env["CI"] = "1"
	config.EnvOverrides = env
	client, e := pirpc.Start(config)
	if e != nil {
		return "", e
	}
	defer client.Close()
	if cfg.Model != "" {
		if _, e = client.Call(ctx, pirpc.NewCommand(pirpc.SetModel, map[string]any{"provider": cfg.Provider, "modelId": cfg.Model})); e != nil {
			return "", e
		}
	}
	if _, e = client.Call(ctx, pirpc.NewCommand(pirpc.Prompt, map[string]any{"message": prompt.User})); e != nil {
		return "", e
	}
	tick := time.NewTicker(15 * time.Millisecond)
	defer tick.Stop()
	lastText, lastError := "", ""
	for {
		for _, event := range client.DrainEvents() {
			switch event.Type {
			case "message_end":
				var message struct {
					Role, StopReason, ErrorMessage string
					Content                        []struct{ Type, Text string }
				}
				if event.DecodeField("message", &message) == nil && message.Role == "assistant" {
					lastText = ""
					for _, block := range message.Content {
						if block.Type == "text" {
							lastText += block.Text
						}
					}
					lastError = ""
					if message.StopReason == "error" || message.StopReason == "aborted" {
						lastError = message.ErrorMessage
						if lastError == "" {
							lastError = message.StopReason
						}
					}
				}
			case "extension_ui_request":
				method := event.String("method")
				if method == "select" || method == "confirm" || method == "input" || method == "editor" {
					_ = client.CancelDialog(ctx, event.ID)
					return "", errors.New("draft generation requested interactive extension input")
				}
			case "agent_end":
				if event.Bool("willRetry") {
					continue
				}
				if lastError != "" {
					return "", errors.New(lastError)
				}
				response, err := client.Call(ctx, pirpc.NewCommand(pirpc.GetLastAssistantText, nil))
				if err != nil {
					if strings.TrimSpace(lastText) != "" {
						return lastText, nil
					}
					return "", err
				}
				var data struct {
					Text string `json:"text"`
				}
				if err = response.DecodeData(&data); err != nil {
					return "", err
				}
				if strings.TrimSpace(data.Text) == "" {
					return "", errors.New("Pi returned no generated text")
				}
				return data.Text, nil
			case "transport_error":
				return "", errors.New(event.String("error"))
			}
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-client.Done():
			return "", client.Err()
		case <-tick.C:
		}
	}
}

// GenerateCommitRPC falls back to an explicitly marked, editable heuristic on
// missing Pi/auth/model failures, but user cancellation never becomes success.
func GenerateCommitRPC(ctx context.Context, g Git, cfg GenerationConfig, stagedOnly bool) (GeneratedText, error) {
	prompt, rows, e := commitGenerationPrompt(ctx, g, stagedOnly, cfg.ContextWindow)
	if e != nil {
		return GeneratedText{}, e
	}
	cfg.Pi.WorkspaceDir = g.Dir
	raw, e := isolatedRPC(ctx, cfg, prompt)
	if e == nil {
		message, parseErr := ParseCommitMessage(raw)
		if parseErr == nil {
			return GeneratedText{Text: message}, nil
		}
		e = parseErr
	}
	if ctx.Err() != nil {
		return GeneratedText{}, ctx.Err()
	}
	return GeneratedText{Text: HeuristicCommitMessage(rows), Fallback: true, Warning: e.Error()}, nil
}
func CommitGenerationPrompt(ctx context.Context, g Git, stagedOnly bool) (GenerationPrompt, []StatusRow, error) {
	return commitGenerationPrompt(ctx, g, stagedOnly, 0)
}
func commitGenerationPrompt(ctx context.Context, g Git, stagedOnly bool, window int) (GenerationPrompt, []StatusRow, error) {
	rows, e := g.Status(ctx)
	if e != nil {
		return GenerationPrompt{}, nil, e
	}
	selected := make([]StatusRow, 0, len(rows))
	for _, row := range rows {
		if !stagedOnly || row.Staged() {
			selected = append(selected, row)
		}
	}
	if len(selected) == 0 {
		return GenerationPrompt{}, nil, errors.New("no changes to describe")
	}
	branch, e := g.CurrentBranch(ctx)
	if e != nil {
		return GenerationPrompt{}, nil, e
	}
	var b strings.Builder
	fmt.Fprintf(&b, "<repository-context>\nRepository name: %s\nBranch name: %s\n</repository-context>\n\n", filepath.Base(g.Dir), branch)
	history, _ := g.History(ctx, HistoryOptions{Limit: 5})
	appendCommitSubjects(&b, "recent-commits", history)
	if name, e := g.UserName(ctx); e == nil && name != "" {
		own, _ := g.History(ctx, HistoryOptions{Limit: 5, Author: name})
		appendCommitSubjects(&b, "user-commits", own)
	}
	b.WriteString("<changes>\n# CHANGED FILES (status and line counts):\n")
	for _, r := range selected {
		badge := r.Index
		adds, dels := r.StagedAdditions, r.StagedDeletions
		if !stagedOnly {
			badge = r.Worktree
			adds += r.UnstagedAdditions
			dels += r.UnstagedDeletions
		}
		fmt.Fprintf(&b, "%c %s (+%d -%d)\n", badge, r.Path, adds, dels)
	}
	budget := newCommitBudget(window, b.Len()+8*1024, len(selected))
	hasHead, e := g.HasCommits(ctx)
	if e != nil {
		return GenerationPrompt{}, nil, e
	}
	for index, r := range selected {
		if e := ctx.Err(); e != nil {
			return GenerationPrompt{}, nil, e
		}
		fmt.Fprintf(&b, "\n<file path=%q>\n<original-code>\n", r.Path)
		originalPath := r.Path
		if r.OrigPath != "" {
			originalPath = r.OrigPath
		}
		if hasHead {
			original, err := g.FileAtRevision(ctx, "HEAD", originalPath)
			if err == nil && utf8.Valid(original) && !strings.ContainsRune(string(original), 0) {
				text := budget.takeOriginal(string(original), len(selected)-index)
				if text != "" {
					b.WriteString(text)
					b.WriteByte('\n')
				} else {
					b.WriteString("// (original code omitted: prompt budget exhausted)\n")
				}

			} else {
				b.WriteString("// (new file or binary original)\n")
			}
		} else {
			b.WriteString("// (new file — no previous version)\n")
		}
		b.WriteString("</original-code>\n<code-changes>\n```diff\n")
		paths := []string{r.Path}
		if r.OrigPath != "" {
			paths = append(paths, r.OrigPath)
		}
		ps, _ := literalPaths(paths)
		args := []string{"diff", "--no-ext-diff", "--no-color"}
		if stagedOnly {
			args = append(args, "--cached")
		} else if hasHead {
			args = append(args, "HEAD")
		}
		patch, e := g.Run(ctx, append(append(args, "--"), ps...)...)
		if e != nil {
			return GenerationPrompt{}, nil, e
		}
		if !stagedOnly && (r.Index == '?' || !hasHead) {
			path := filepath.Join(g.Dir, r.Path)
			raw, err := readGenerationFile(path, MaxGenerationDiffBytes)
			if errors.Is(err, os.ErrNotExist) {
				raw = "(file removed)"
				err = nil
			}
			if err != nil {
				return GenerationPrompt{}, nil, err
			}
			patch = []byte("New untracked file: " + r.Path + "\n" + raw)
		}
		excerpt := budget.takeDiff(string(patch), len(selected)-index)
		if excerpt == "" && len(patch) > 0 {
			b.WriteString("(diff omitted: shared prompt budget exhausted)")
		} else {
			b.WriteString(excerpt)
		}
		b.WriteString("\n```\n</code-changes>\n</file>\n")
	}
	if stagedOnly {
		unstaged, e := g.UnstagedPatch(ctx)
		if e != nil {
			return GenerationPrompt{}, nil, e
		}
		if strings.TrimSpace(unstaged) != "" {
			b.WriteString("# UNSTAGED CHANGES (context only, not part of this commit):\n```diff\n" + budget.takeDiff(unstaged, 1) + "\n```\n")
		}
	}
	reminder := "</changes>\n\n<reminder>Describe only the CODE CHANGES, not context-only changes. Follow the recent author's/repository's conventions without copying their subjects. Return one text code fence containing a concise conventional-commit subject and optional past-tense bullet body. Do not add issue references, authors, explanations or unsupported claims.</reminder>"
	system := "You help a developer write a precise Git commit message. Analyze the code changes using original code as context. Identify why the changes were made; use recent user and repository commits only for writing conventions. Never invent facts. Return only one fenced text block with the commit message, no commentary."
	// Fixed metadata and markers also count. Oversized metadata is explicitly
	// truncated; never silently send a multi-megabyte prompt from many paths.
	user := commitBound(b.String(), budget.total-len(system)-len(reminder)) + reminder
	return GenerationPrompt{System: system, User: user}, selected, nil
}
func appendCommitSubjects(b *strings.Builder, tag string, entries []CommitEntry) {
	if len(entries) == 0 {
		return
	}
	fmt.Fprintf(b, "<%s>\n", tag)
	for _, entry := range entries {
		b.WriteString("- " + entry.Subject + "\n")
	}
	fmt.Fprintf(b, "</%s>\n\n", tag)
}

var numberedBullet = regexp.MustCompile(`^[0-9]+[.)]\s*`)

func stripBullet(s string) string {
	s = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(s), "-*•–—"))
	return numberedBullet.ReplaceAllString(s, "")
}
func extractGeneratedText(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	for _, marker := range []string{"```text", "```markdown", "```md", "```"} {
		index := strings.Index(raw, marker)
		if index < 0 || (index > 0 && raw[index-1] != '\n') {
			continue
		}
		rest := raw[index+len(marker):]
		if strings.HasPrefix(rest, "\r\n") {
			rest = rest[2:]
		} else if strings.HasPrefix(rest, "\n") {
			rest = rest[1:]
		} else {
			continue
		}
		if end := strings.Index(rest, "\n```"); end >= 0 {
			body := strings.TrimSpace(rest[:end])
			if body != "" {
				return body, nil
			}
		}
	}
	raw = strings.TrimSpace(strings.Trim(raw, "\""))
	if raw == "" || strings.HasPrefix(raw, "```") {
		return "", errors.New("generated response was empty")
	}
	return raw, nil
}
func ParseCommitMessage(raw string) (string, error) {
	text, e := extractGeneratedText(raw)
	if e != nil {
		return "", e
	}
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		line = stripBullet(line)
		if line == "" {
			continue
		}
		if len(lines) == 0 {
			lines = append(lines, strings.TrimRight(line, "."))
			continue
		}
		r, n := utf8.DecodeRuneInString(line)
		line = string(unicode.ToUpper(r)) + line[n:]
		if !strings.ContainsAny(line[len(line)-1:], ".!?") {
			line += "."
		}
		lines = append(lines, "- "+line)
	}
	if len(lines) == 0 {
		return "", errors.New("empty commit message")
	}
	if len(lines) == 1 {
		return lines[0], nil
	}
	return lines[0] + "\n\n" + strings.Join(lines[1:], "\n"), nil
}
func humanizeFile(path string) string {
	return strings.NewReplacer("_", " ", "-", " ").Replace(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
}
func joinNames(names []string) string {
	switch len(names) {
	case 0:
		return "files"
	case 1:
		return names[0]
	case 2:
		return names[0] + " and " + names[1]
	}
	return strings.Join(names[:len(names)-1], ", ") + ", and " + names[len(names)-1]
}
func HeuristicCommitMessage(rows []StatusRow) string {
	var added, removed, changed, names []string
	seen := map[string]bool{}
	for _, r := range rows {
		name := humanizeFile(r.Path)
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
		badge := r.Index
		if badge == ' ' {
			badge = r.Worktree
		}
		switch badge {
		case 'A', '?':
			added = append(added, name)
		case 'D':
			removed = append(removed, name)
		default:
			changed = append(changed, name)
		}
	}
	kind, verb := "chore", "update"
	if len(added) > 0 {
		kind = "feat"
	}
	if len(rows) > 0 && len(removed) == len(rows) {
		verb = "remove"
	} else if len(rows) > 0 && len(added) == len(rows) {
		verb = "add"
	}
	scope := ""
	if len(rows) > 0 {
		parts := strings.Split(rows[0].Path, "/")
		n := len(parts) - 1
		for _, r := range rows[1:] {
			p := strings.Split(r.Path, "/")
			i := 0
			for i < n && i+1 < len(p) && p[i] == parts[i] {
				i++
			}
			n = i
		}
		if n > 0 {
			scope = "(" + parts[n-1] + ")"
		}
	}
	description := joinNames(names[:min(2, len(names))])
	subject := kind + scope + ": " + verb + " " + description
	var body []string
	if len(changed) > 0 {
		body = append(body, "- Updated "+joinNames(changed)+".")
	}
	if len(added) > 0 {
		body = append(body, "- Added "+joinNames(added)+".")
	}
	if len(removed) > 0 {
		body = append(body, "- Removed "+joinNames(removed)+".")
	}
	if len(body) == 0 {
		return subject
	}
	return subject + "\n\n" + strings.Join(body, "\n")
}

func GenerateDraftRPC(ctx context.Context, g Git, cfg GenerationConfig, request DraftRequest) (Draft, error) {
	prompt, fallback, e := DraftGenerationPrompt(ctx, g, request)
	if e != nil {
		return Draft{}, e
	}
	cfg.Pi.WorkspaceDir = g.Dir
	raw, e := isolatedRPC(ctx, cfg, prompt)
	if e == nil {
		draft, parseErr := ParseDraft(raw)
		if parseErr == nil {
			if request.Template != nil && request.Template.TitlePrefix != "" && !strings.HasPrefix(draft.Title, request.Template.TitlePrefix) {
				draft.Title = request.Template.TitlePrefix + draft.Title
			}
			return draft, nil
		}
		e = parseErr
	}
	if ctx.Err() != nil {
		return Draft{}, ctx.Err()
	}
	fallback.Fallback = true
	fallback.Warning = e.Error()
	return fallback, nil
}
func DraftGenerationPrompt(ctx context.Context, g Git, r DraftRequest) (GenerationPrompt, Draft, error) {
	if r.Kind != "issue" && r.Kind != "pull" {
		return GenerationPrompt{}, Draft{}, errors.New("draft kind must be issue or pull")
	}
	branch, e := g.CurrentBranch(ctx)
	if e != nil {
		return GenerationPrompt{}, Draft{}, e
	}
	args := []string{"log", "--no-decorate", "--format=%s", "-n", "30"}
	if r.Base != "" {
		if e := validRevision(r.Base); e != nil {
			return GenerationPrompt{}, Draft{}, e
		}
		args = append(args, r.Base+"..HEAD")
	}
	raw, e := g.Run(ctx, args...)
	if e != nil {
		if ctx.Err() != nil {
			return GenerationPrompt{}, Draft{}, ctx.Err()
		}
		raw = nil
	}
	commits := nonemptyLines(strings.TrimSpace(string(raw)))
	statArgs := []string{"diff", "--stat"}
	if r.Base != "" {
		statArgs = append(statArgs, r.Base+"..HEAD")
	} else {
		if ok, _ := g.HasCommits(ctx); ok {
			statArgs = append(statArgs, "HEAD")
		}
	}
	stat, e := g.Run(ctx, statArgs...)
	if e != nil && ctx.Err() != nil {
		return GenerationPrompt{}, Draft{}, ctx.Err()
	}
	var b strings.Builder
	fmt.Fprintf(&b, "<repository-context>\nRepository: %s\nBranch: %s\nTarget base: %s\n</repository-context>\n<user-notes>\n%s\n</user-notes>\n<commits>\n%s\n</commits>\n<changed-files>\n%s\n</changed-files>\n", filepath.Base(g.Dir), branch, r.Base, boundedText(r.Notes, MaxDraftContextBytes), boundedText(strings.Join(commits, "\n"), MaxDraftContextBytes), boundedText(string(stat), MaxDraftContextBytes))
	if r.Template != nil {
		fmt.Fprintf(&b, "<template name=%q>\n%s\n</template>\n", r.Template.Label(), boundedText(r.Template.Body, MaxDraftContextBytes))
	}
	b.WriteString("Return a single text code fence: one title line, a blank line, then a Markdown body.")
	system := "Draft a concise GitHub " + r.Kind + " from user notes and repository context. Do not invent facts or test outcomes. Use short headings and bullets. Explain why and what. A pull request may include Summary and Testing only when testing evidence exists. An issue may include What happened and Expected when notes support them. Preserve every supplied template heading and unchecked checkbox exactly; fill matching sections, remove guidance comments, and do not check boxes without evidence. Return only the requested text fence."
	fallback := heuristicDraft(r, branch, commits, string(stat))
	return GenerationPrompt{system, b.String()}, fallback, nil
}
func ParseDraft(raw string) (Draft, error) {
	text, e := extractGeneratedText(raw)
	if e != nil {
		return Draft{}, e
	}
	lines := strings.Split(text, "\n")
	title := ""
	i := 0
	for ; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			title = cleanDraftTitle(lines[i])
			i++
			break
		}
	}
	if title == "" {
		return Draft{}, errors.New("generated draft has no title")
	}
	return Draft{Title: title, Body: strings.TrimSpace(strings.Join(lines[i:], "\n"))}, nil
}
func cleanDraftTitle(s string) string {
	s = strings.Trim(strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(s), "-*•#")), "\"`'")
	s = strings.TrimPrefix(strings.TrimPrefix(s, "Title:"), "title:")
	return boundedText(strings.TrimSpace(strings.TrimRight(strings.TrimSpace(s), ".:;")), 120)
}
func heuristicDraft(r DraftRequest, branch string, commits []string, stats string) Draft {
	notes := strings.TrimSpace(r.Notes)
	title := ""
	if notes != "" {
		title = cleanDraftTitle(strings.Split(notes, "\n")[0])
	} else if len(commits) > 0 {
		title = cleanDraftTitle(commits[0])
	} else {
		title = strings.NewReplacer("-", " ", "_", " ").Replace(filepath.Base(branch))
	}
	if title == "" {
		title = "Describe the requested change"
	}
	var b strings.Builder
	heading := "What happened"
	if r.Kind == "pull" {
		heading = "Summary"
	}
	b.WriteString("## " + heading + "\n\n")
	for _, c := range commits[:min(10, len(commits))] {
		b.WriteString("- " + c + "\n")
	}
	if len(commits) == 0 && notes != "" {
		b.WriteString(notes + "\n")
	}
	if strings.TrimSpace(stats) != "" {
		b.WriteString("\n## Changed files\n\n" + boundedText(strings.TrimSpace(stats), MaxDraftContextBytes))
	}
	body := strings.TrimSpace(b.String())
	if r.Template != nil {
		seed := notes
		if seed == "" {
			seed = strings.Join(commits[:min(10, len(commits))], "\n")
		}
		body = fillTemplateSection(r.Template.Body, seed)
		if r.Template.TitlePrefix != "" && !strings.HasPrefix(title, r.Template.TitlePrefix) {
			title = r.Template.TitlePrefix + title
		}
	}
	return Draft{Title: title, Body: body}
}
func fillTemplateSection(template, text string) string {
	template = strings.TrimSpace(template)
	text = strings.TrimSpace(text)
	if text == "" || template == "" {
		return template
	}
	lines := strings.Split(template, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimLeft(line, " \t"), "#") {
			return strings.Join(lines[:i+1], "\n") + "\n\n" + text + "\n\n" + strings.TrimLeft(strings.Join(lines[i+1:], "\n"), "\n")
		}
	}
	return text + "\n\n" + template
}

func readGenerationFile(path string, limit int) (string, error) {
	info, e := os.Lstat(path)
	if e != nil {
		return "", e
	}
	if !info.Mode().IsRegular() {
		return "(non-regular file content omitted)", nil
	}
	file, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer file.Close()
	raw, e := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	if e != nil {
		return "", e
	}
	if !utf8.Valid(raw) || strings.ContainsRune(string(raw), 0) {
		return "(binary file)", nil
	}
	return boundedText(string(raw), limit), nil
}
