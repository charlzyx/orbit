package workspace

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/charlzyx/orbit/internal/pirpc"
)

func generationConfig(t *testing.T, mode, text string) GenerationConfig {
	t.Helper()
	return GenerationConfig{Pi: pirpc.Config{Executable: os.Args[0], PrefixArgs: []string{"-test.run=TestGenerationRPCProcess", "--"}, EnvOverrides: map[string]string{"ORBIT_GENERATION_HELPER": "1", "ORBIT_GENERATION_MODE": mode, "ORBIT_GENERATION_TEXT": text, "ORBIT_GENERATION_ARGS": filepath.Join(t.TempDir(), "args.json")}, ShutdownTimeout: 50 * time.Millisecond}, Provider: "test-provider", Model: "test-model", Timeout: 2 * time.Second}
}
func TestGenerationRPCProcess(t *testing.T) {
	if os.Getenv("ORBIT_GENERATION_HELPER") != "1" {
		return
	}
	raw, _ := json.Marshal(os.Args[1:])
	_ = os.WriteFile(os.Getenv("ORBIT_GENERATION_ARGS"), raw, 0600)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4096), 8<<20)
	encoder := json.NewEncoder(os.Stdout)
	for scanner.Scan() {
		var request map[string]any
		if json.Unmarshal(scanner.Bytes(), &request) != nil {
			os.Exit(3)
		}
		kind, _ := request["type"].(string)
		response := map[string]any{"type": "response", "id": request["id"], "command": kind, "success": true}
		switch kind {
		case "prompt":
			if file := os.Getenv("ORBIT_GENERATION_PROMPT"); file != "" {
				_ = os.WriteFile(file, scanner.Bytes(), 0600)
			}
			if os.Getenv("ORBIT_GENERATION_MODE") == "reject" {
				response["success"] = false
				response["error"] = "no credentials"
			}
			_ = encoder.Encode(response)
			if os.Getenv("ORBIT_GENERATION_MODE") != "reject" && os.Getenv("ORBIT_GENERATION_MODE") != "hang" {
				_ = encoder.Encode(map[string]any{"type": "agent_end"})
			}
			continue
		case "set_model":
			if request["provider"] != "test-provider" || request["modelId"] != "test-model" {
				response["success"] = false
				response["error"] = "wrong selected model"
			}
		case "get_last_assistant_text":
			response["data"] = map[string]any{"text": os.Getenv("ORBIT_GENERATION_TEXT")}
		}
		_ = encoder.Encode(response)
	}
	os.Exit(0)
}
func TestCommitRPCIsolatedAndNormalized(t *testing.T) {
	ctx := context.Background()
	g := testGit(t)
	testCommit(t, g, "src/test.go", "before\n", "style: earlier")
	writeGitFile(t, g, "src/test.go", "after\n")
	requireNoError(t, g.Stage(ctx, "src/test.go"))
	cfg := generationConfig(t, "ok", "Preface\n```text\nfeat(src): add native support.\n\n* added typed operations\n2) preserved errors\n```\nTrailing")
	result, e := GenerateCommitRPC(ctx, g, cfg, true)
	requireNoError(t, e)
	if result.Fallback || result.Text != "feat(src): add native support\n\n- Added typed operations.\n- Preserved errors." {
		t.Fatal(result)
	}
	raw, e := os.ReadFile(cfg.Pi.EnvOverrides["ORBIT_GENERATION_ARGS"])
	requireNoError(t, e)
	var args []string
	requireNoError(t, json.Unmarshal(raw, &args))
	joined := strings.Join(args, " ")
	for _, flag := range []string{"--mode rpc", "--no-session", "--no-tools", "--no-skills", "--no-context-files", "--no-approve", "--system-prompt"} {
		if !strings.Contains(joined, flag) {
			t.Fatalf("missing %s: %v", flag, args)
		}
	}
	for _, arg := range args {
		if arg == "-p" || arg == "--approve" {
			t.Fatal("unsafe generation flag", arg)
		}
	}
	rows, e := g.Status(ctx)
	requireNoError(t, e)
	if len(rows) != 1 || !rows[0].Staged() {
		t.Fatal("generation modified index/worktree", rows)
	}
}
func TestDraftRPCFallbackTimeoutAndCancellation(t *testing.T) {
	ctx := context.Background()
	g := testGit(t)
	testCommit(t, g, "a", "a", "Initial")
	template := RepositoryTemplate{TitlePrefix: "[Bug] ", Body: "## Summary\n\n## Checks\n- [ ] Verified"}
	request := DraftRequest{Kind: "issue", Notes: "Upload fails\nFails on retry", Template: &template}
	cfg := generationConfig(t, "ok", "```markdown\nUpload fails\n\n## Summary\n\nRetry fails\n\n## Checks\n- [ ] Verified\n```")
	draft, e := GenerateDraftRPC(ctx, g, cfg, request)
	requireNoError(t, e)
	if draft.Fallback || draft.Title != "[Bug] Upload fails" || !strings.Contains(draft.Body, "- [ ] Verified") {
		t.Fatal(draft)
	}
	for _, mode := range []string{"reject", "hang"} {
		cfg = generationConfig(t, mode, "")
		cfg.Timeout = 60 * time.Millisecond
		draft, e = GenerateDraftRPC(ctx, g, cfg, request)
		requireNoError(t, e)
		if !draft.Fallback || draft.Warning == "" || !strings.Contains(draft.Body, "Fails on retry") || !strings.Contains(draft.Body, "- [ ] Verified") {
			t.Fatal(mode, draft)
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	cfg = generationConfig(t, "ok", "not used")
	_, e = GenerateDraftRPC(cancelled, g, cfg, request)
	if !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, e = os.Stat(cfg.Pi.EnvOverrides["ORBIT_GENERATION_ARGS"]); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("cancelled generation spawned Pi")
	}
}
func TestCommitContextBudgetsAndNoMutation(t *testing.T) {
	ctx := context.Background()
	g := testGit(t)
	for _, name := range []string{"a", "b", "c", "d", "e"} {
		writeGitFile(t, g, name, strings.Repeat("original context line\n", 4000))
	}
	_, e := g.Commit(ctx, "repository writing convention", true)
	requireNoError(t, e)
	for _, name := range []string{"a", "b", "c", "d", "e"} {
		writeGitFile(t, g, name, strings.Repeat("original context line\n", 4000)+"changed line\n")
	}
	requireNoError(t, g.StageAll(ctx))
	writeGitFile(t, g, "a", strings.Repeat("original context line\n", 4000)+"changed line\nunstaged context only\n")
	prompt, rows, e := CommitGenerationPrompt(ctx, g, true)
	requireNoError(t, e)
	if len(rows) != 5 || !strings.Contains(prompt.User, "<user-commits>") || !strings.Contains(prompt.User, "repository writing convention") || !strings.Contains(prompt.User, "unstaged context only") || !strings.Contains(prompt.User, "prompt budget exhausted") {
		t.Fatal("missing original/history/unstaged budget context")
	}
	if len(prompt.User) > MaxGenerationTotalOriginalBytes+MaxGenerationDiffBytes+10*1024 {
		t.Fatalf("original code prompt budget unbounded %d", len(prompt.User))
	}
	if !utf8.ValidString(boundedText(strings.Repeat("你好", 10), 7)) {
		t.Fatal("invalid UTF-8 truncation")
	}
	state, e := g.Status(ctx)
	requireNoError(t, e)
	if !state[0].Staged() || !state[0].Unstaged() {
		t.Fatal("prompt collection mutated files")
	}
}
func TestCommitUnbornAndUntrackedContext(t *testing.T) {
	ctx := context.Background()
	g := testGit(t)
	writeGitFile(t, g, "new.txt", "new untracked content")
	prompt, _, e := CommitGenerationPrompt(ctx, g, false)
	requireNoError(t, e)
	if !strings.Contains(prompt.User, "new untracked content") {
		t.Fatal(prompt.User)
	}
	requireNoError(t, g.Stage(ctx, "new.txt"))
	prompt, _, e = CommitGenerationPrompt(ctx, g, false)
	requireNoError(t, e)
	if !strings.Contains(prompt.User, "new untracked content") {
		t.Fatal("unborn staged contents omitted")
	}
	cfg := GenerationConfig{Pi: pirpc.Config{Executable: filepath.Join(g.Dir, "missing-pi")}}
	result, e := GenerateCommitRPC(ctx, g, cfg, true)
	requireNoError(t, e)
	if !result.Fallback || !strings.HasPrefix(result.Text, "feat:") || result.Warning == "" {
		t.Fatal(result)
	}
}
func TestGenerationParsersAndHeuristics(t *testing.T) {
	for _, raw := range []string{"", "```text\n```"} {
		if _, e := ParseDraft(raw); e == nil {
			t.Fatal("empty draft accepted")
		}
		if _, e := ParseCommitMessage(raw); e == nil {
			t.Fatal("empty commit accepted")
		}
	}
	draft, e := ParseDraft("Preamble\n```text\nTitle: Fix retries.\n\nBody\n```\nTrailing")
	requireNoError(t, e)
	if draft.Title != "Fix retries" || draft.Body != "Body" {
		t.Fatal(draft)
	}
	message := HeuristicCommitMessage([]StatusRow{{Path: "src/git_ops.go", Index: 'A'}, {Path: "src/git_test.go", Index: 'M'}})
	if !strings.HasPrefix(message, "feat(src):") || !strings.Contains(message, "Added git ops") || !strings.Contains(message, "Updated git test") {
		t.Fatal(message)
	}
	if got := HeuristicCommitMessage([]StatusRow{{Path: "old.txt", Index: 'D'}}); !strings.HasPrefix(got, "chore: remove") {
		t.Fatal(got)
	}
}
func TestRepositoryTemplatesLookupAndForms(t *testing.T) {
	root := t.TempDir()
	write := func(path, body string) {
		t.Helper()
		p := filepath.Join(root, path)
		requireNoError(t, os.MkdirAll(filepath.Dir(p), 0700))
		requireNoError(t, os.WriteFile(p, []byte(body), 0600))
	}
	if len(IssueTemplates(root)) != 0 {
		t.Fatal("missing templates not empty")
	}
	write(".github/PULL_REQUEST_TEMPLATE.md", "## Summary\n\n- [ ] Tests")
	pulls := PullTemplates(root)
	if len(pulls) != 1 || pulls[0].Path != ".github/PULL_REQUEST_TEMPLATE.md" {
		t.Fatal(pulls)
	}
	write(".github/ISSUE_TEMPLATE.md", "legacy")
	write(".github/ISSUE_TEMPLATE/bug.md", "---\nname: 'Bug report'\ntitle: \"bug: \"\nlabels: [bug, ui]\n---\n## What happened\n")
	write(".github/ISSUE_TEMPLATE/config.yml", "blank_issues_enabled: false")
	write(".github/ISSUE_TEMPLATE/form.yml", "name: Feature request\ntitle: '[Feature] '\nlabels: enhancement\nbody:\n  - type: markdown\n    attributes:\n      value: |\n        Helpful guidance\n  - type: textarea\n    attributes:\n      label: Description\n      description: Explain why\n      placeholder: |\n        First line\n        Second line\n      render: shell\n  - type: dropdown\n    attributes:\n      label: Platform\n      options:\n        - macOS\n        - Linux\n  - type: checkboxes\n    attributes:\n      label: Checklist\n      options:\n        - label: I searched existing issues\n          required: true\n")
	issues := IssueTemplates(root)
	if len(issues) != 2 || issues[0].Name != "Bug report" || issues[0].TitlePrefix != "bug: " || !reflect.DeepEqual(issues[0].Labels, []string{"bug", "ui"}) {
		t.Fatal(issues)
	}
	body := issues[1].Body
	for _, text := range []string{"<!--\nHelpful guidance\n-->", "### Description", "First line\nSecond line", "```shell", "one of: macOS, Linux", "- [ ] I searched existing issues"} {
		if !strings.Contains(body, text) {
			t.Fatalf("missing %q in %s", text, body)
		}
	}
	write(".github/ISSUE_TEMPLATE/too-large.md", strings.Repeat("x", MaxTemplateBytes+1))
	if len(IssueTemplates(root)) != 2 {
		t.Fatal("oversized template read")
	}
	outside := filepath.Join(t.TempDir(), "outside.md")
	requireNoError(t, os.WriteFile(outside, []byte("outside"), 0600))
	if os.Symlink(outside, filepath.Join(root, ".github/ISSUE_TEMPLATE/link.md")) == nil && len(IssueTemplates(root)) != 2 {
		t.Fatal("outside symlink read")
	}
}

func TestReportDraftRPCNotesOnlyFallbackAndCancellation(t *testing.T) {
	ctx := context.Background()
	for _, kind := range []string{"bug", "feature", "other"} {
		prompt, fallback, err := ReportGenerationPrompt(kind, "Save crashes.\n\n1. Open app\n2. Save")
		requireNoError(t, err)
		if fallback.Title != "Save crashes" || !strings.Contains(prompt.User, "1. Open app") || strings.Contains(prompt.User, "repository") || !strings.Contains(prompt.System, "no headings") {
			t.Fatal(prompt, fallback)
		}
	}
	if _, _, err := ReportGenerationPrompt("invalid", "text"); err == nil {
		t.Fatal("invalid report kind accepted")
	}
	cfg := generationConfig(t, "ok", "```text\nSave crashes\n\nThe app closes on Save.\n\n1. Open app\n2. Save\n```")
	cfg.Pi.WorkspaceDir = t.TempDir()
	requireNoError(t, os.WriteFile(filepath.Join(cfg.Pi.WorkspaceDir, "AGENTS.md"), []byte("WORKSPACE_SECRET_SENTINEL"), 0600))
	capture := filepath.Join(t.TempDir(), "prompt.json")
	cfg.Pi.EnvOverrides["ORBIT_GENERATION_PROMPT"] = capture
	result, err := GenerateReportDraftRPC(ctx, cfg, "bug", "Save crashes after opening")
	requireNoError(t, err)
	if result.Fallback || result.Title != "Save crashes" {
		t.Fatal(result)
	}
	raw, err := os.ReadFile(capture)
	requireNoError(t, err)
	if strings.Contains(string(raw), "WORKSPACE_SECRET_SENTINEL") || strings.Contains(string(raw), "repository-context") || !strings.Contains(string(raw), "Save crashes after opening") {
		t.Fatal("report sampled unrelated context", string(raw))
	}
	cfg = generationConfig(t, "reject", "")
	result, err = GenerateReportDraftRPC(ctx, cfg, "feature", "More themes.\nPlease add custom colors")
	requireNoError(t, err)
	if !result.Fallback || result.Warning == "" || result.Title != "More themes" || result.Body != "More themes.\nPlease add custom colors" {
		t.Fatal(result)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	cfg = generationConfig(t, "ok", "unused")
	_, err = GenerateReportDraftRPC(cancelled, cfg, "other", "Report")
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err = os.Stat(cfg.Pi.EnvOverrides["ORBIT_GENERATION_ARGS"]); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("cancelled report generation spawned Pi")
	}
	_, fallback, err := ReportGenerationPrompt("bug", strings.Repeat("界", 100)+".\nbody")
	requireNoError(t, err)
	if len([]rune(fallback.Title)) != 72 {
		t.Fatal("fallback title not unicode safe", fallback.Title)
	}
}
