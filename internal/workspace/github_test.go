package workspace

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func fakeGitHub(output string) (GitHub, *[][]string) {
	calls := [][]string{}
	h := GitHub{runner: func(ctx context.Context, args ...string) ([]byte, error) {
		calls = append(calls, append([]string(nil), args...))
		return []byte(output), nil
	}}
	return h, &calls
}
func TestGitHubProbeAndVersions(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		version                    string
		installed, auth, supported bool
		failAt                     int
	}{{"gh version 2.20.0", true, true, true, 0}, {"gh version 2.19.9", true, true, false, 0}, {"gh version 3.0.0", true, true, true, 0}, {"gh nightly", true, true, true, 0}, {"", false, false, false, 1}, {"gh version 2.40.0", true, false, true, 2}} {
		t.Run(tc.version+string(rune(tc.failAt+'0')), func(t *testing.T) {
			calls := 0
			h := GitHub{runner: func(_ context.Context, args ...string) ([]byte, error) {
				calls++
				if calls == tc.failAt {
					return nil, errors.New("not available")
				}
				return []byte(tc.version), nil
			}}
			s, e := h.Probe(ctx)
			requireNoError(t, e)
			if s.Installed != tc.installed || s.Authenticated != tc.auth || s.Supported != tc.supported {
				t.Fatal(s)
			}
		})
	}
	for _, v := range []string{"gh version 2.40.1 (date)", "v2.40.1-beta"} {
		n, ok := ParseGitHubVersion(v)
		if !ok || n != [3]int{2, 40, 1} {
			t.Fatal(n, ok)
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	h, calls := fakeGitHub("{}")
	if _, e := h.Probe(ctx); !errors.Is(e, context.Canceled) || len(*calls) != 0 {
		t.Fatal(e, *calls)
	}
}
func TestGitHubIssuesAndUserParsing(t *testing.T) {
	ctx := context.Background()
	h, calls := fakeGitHub(`[{"number":12,"title":"Issue","state":"OPEN","author":null,"labels":[{"name":"bug","color":"ff0000"}],"comments":[{"author":{"login":"dev"},"body":"reply"}],"futureField":true}]`)
	issues, e := h.Issues(ctx, IssueFilter{State: "closed", Label: "bug", Assignee: "@me", Search: "a b", Limit: 25})
	requireNoError(t, e)
	if len(issues) != 1 || issues[0].Author != nil || issues[0].Comments[0].Author.Login != "dev" {
		t.Fatal(issues)
	}
	want := []string{"issue", "list", "--json", issueListFields, "--limit", "25", "--state", "closed", "--label", "bug", "--assignee", "@me", "--search", "a b"}
	if !reflect.DeepEqual((*calls)[0], want) {
		t.Fatal(*calls)
	}
	h, calls = fakeGitHub(`{"number":12,"title":"Issue","body":"full","state":"OPEN"}`)
	issue, e := h.Issue(ctx, 12)
	requireNoError(t, e)
	if issue.Body != "full" || !issue.IsOpen() {
		t.Fatal(issue)
	}
	h, _ = fakeGitHub(`{"login":"dev","name":null,"avatar_url":"https://example.invalid/a"}`)
	user, e := h.CurrentUser(ctx)
	requireNoError(t, e)
	if user.Login != "dev" || user.AvatarURL == "" {
		t.Fatal(user)
	}
	h, _ = fakeGitHub(`[{"name":"bug","color":"f00","description":null}]`)
	labels, e := h.Labels(ctx)
	requireNoError(t, e)
	if len(labels) != 1 {
		t.Fatal(labels)
	}
	h, _ = fakeGitHub("not JSON")
	if _, e := h.Issues(ctx, IssueFilter{}); e == nil {
		t.Fatal("malformed JSON succeeded")
	}
}
func TestGitHubIssueMutationArguments(t *testing.T) {
	ctx := context.Background()
	h, calls := fakeGitHub("https://github.com/example/repo/issues/1\n")
	out, e := h.CreateIssue(ctx, CreateIssueOptions{Repo: "example/repo", Title: "quoted ' title", Body: "multiline\nbody", Labels: []string{"bug", "urgent"}, Assignees: []string{"dev"}, Attachments: []string{"/tmp/a b.png"}})
	requireNoError(t, e)
	if strings.HasSuffix(out, "\n") {
		t.Fatal("URL untrimmed")
	}
	want := []string{"issue", "create", "--repo", "example/repo", "--title", "quoted ' title", "--body", "multiline\nbody", "--attach", "/tmp/a b.png", "--label", "bug", "--label", "urgent", "--assignee", "dev"}
	if !reflect.DeepEqual((*calls)[0], want) {
		t.Fatal(*calls)
	}
	_, e = h.CommentIssue(ctx, 1, "body")
	requireNoError(t, e)
	_, e = h.EditIssueLabels(ctx, 1, []string{"x"}, []string{"y"})
	requireNoError(t, e)
	_, e = h.CloseIssue(ctx, 1, "not planned")
	requireNoError(t, e)
	_, e = h.ReopenIssue(ctx, 1)
	requireNoError(t, e)
	if !reflect.DeepEqual((*calls)[2], []string{"issue", "edit", "1", "--add-label", "x", "--remove-label", "y"}) {
		t.Fatal(*calls)
	}
	if !reflect.DeepEqual((*calls)[3], []string{"issue", "close", "1", "--reason", "not planned"}) {
		t.Fatal(*calls)
	}
	if !reflect.DeepEqual((*calls)[4], []string{"issue", "reopen", "1"}) {
		t.Fatal(*calls)
	}
}
func TestGitHubPullParsingAndChecks(t *testing.T) {
	ctx := context.Background()
	raw := `{"number":7,"title":"PR","state":"OPEN","isDraft":true,"reviewDecision":null,"mergeable":"CONFLICTING","mergeStateStatus":"DIRTY","headRefName":"feature","baseRefName":"main","statusCheckRollup":[{"name":"CI","status":"IN_PROGRESS","conclusion":""},{"context":"external","state":"SUCCESS","targetUrl":"https://example.invalid"}],"comments":[],"reviews":[{"state":"APPROVED","author":null}],"commits":[{"oid":"abc","messageHeadline":"commit"}],"files":[{"path":"a","changeType":"MODIFIED","additions":2,"deletions":1}]}`
	h, _ := fakeGitHub(raw)
	pr, e := h.Pull(ctx, 7)
	requireNoError(t, e)
	if pr.HeadRef != "feature" || pr.BaseRef != "main" || len(pr.Files) != 1 || pr.Files[0].Additions != 2 || !pr.IsOpen() || pr.IsMerged() {
		t.Fatal(pr)
	}
	bucket, has := pr.ChecksSummary()
	if !has || bucket != CheckPending {
		t.Fatal(bucket, has)
	}
	pr.Checks = append(pr.Checks, GitHubCheck{Conclusion: "FAILURE"})
	bucket, has = pr.ChecksSummary()
	if !has || bucket != CheckFailed {
		t.Fatal(bucket)
	}
	for _, tc := range []struct {
		c    GitHubCheck
		want CheckBucket
	}{{GitHubCheck{Bucket: "pass"}, CheckPassed}, {GitHubCheck{Bucket: "fail"}, CheckFailed}, {GitHubCheck{Bucket: "cancel"}, CheckSkipped}, {GitHubCheck{Conclusion: "ACTION_REQUIRED"}, CheckFailed}, {GitHubCheck{Conclusion: "CANCELLED"}, CheckSkipped}, {GitHubCheck{State: "QUEUED"}, CheckPending}, {GitHubCheck{}, CheckUnknown}} {
		if got := tc.c.BucketKind(); got != tc.want {
			t.Fatal(got, tc.want)
		}
	}
	if _, has := (GitHubPull{}).ChecksSummary(); has {
		t.Fatal("empty checks reported CI")
	}
	c := GitHubCheck{Context: "name", TargetURL: "link"}
	if c.Label() != "name" || c.URL() != "link" {
		t.Fatal(c)
	}
	h, calls := fakeGitHub("[" + raw + "]")
	prs, e := h.Pulls(ctx, PullFilter{State: "merged", Base: "main", Author: "dev", Search: "label:x"})
	requireNoError(t, e)
	if len(prs) != 1 || !reflect.DeepEqual((*calls)[0], []string{"pr", "list", "--json", pullListFields, "--limit", "50", "--state", "merged", "--base", "main", "--author", "dev", "--search", "label:x"}) {
		t.Fatal(prs, *calls)
	}
}
func TestGitHubPullMutations(t *testing.T) {
	ctx := context.Background()
	h, calls := fakeGitHub("done")
	_, e := h.CreatePull(ctx, CreatePullOptions{Title: "title", Body: "body", Base: "main", Draft: true, Labels: []string{"bug"}})
	requireNoError(t, e)
	_, e = h.CommentPull(ctx, 4, "comment")
	requireNoError(t, e)
	_, e = h.ReviewPull(ctx, 4, ReviewApprove, "")
	requireNoError(t, e)
	_, e = h.ReviewPull(ctx, 4, ReviewRequestChanges, "fix it")
	requireNoError(t, e)
	_, e = h.MergePull(ctx, 4, PullSquash, true)
	requireNoError(t, e)
	_, e = h.ClosePull(ctx, 4)
	requireNoError(t, e)
	_, e = h.ReopenPull(ctx, 4)
	requireNoError(t, e)
	wants := [][]string{{"pr", "create", "--title", "title", "--body", "body", "--base", "main", "--draft", "--label", "bug"}, {"pr", "comment", "4", "--body", "comment"}, {"pr", "review", "4", "--approve"}, {"pr", "review", "4", "--request-changes", "--body", "fix it"}, {"pr", "merge", "4", "--squash", "--delete-branch"}, {"pr", "close", "4"}, {"pr", "reopen", "4"}}
	if !reflect.DeepEqual(*calls, wants) {
		t.Fatal(*calls)
	}
}
func TestGitHubValidationPreventsInvocation(t *testing.T) {
	ctx := context.Background()
	h, calls := fakeGitHub("{}")
	invalid := []func() error{func() error { _, e := h.Issues(ctx, IssueFilter{State: "merged"}); return e }, func() error { _, e := h.Pulls(ctx, PullFilter{State: "bad"}); return e }, func() error { _, e := h.Issue(ctx, 0); return e }, func() error { _, e := h.CreateIssue(ctx, CreateIssueOptions{}); return e }, func() error { _, e := h.CreatePull(ctx, CreatePullOptions{}); return e }, func() error { _, e := h.CommentPull(ctx, 2, " "); return e }, func() error { _, e := h.CloseIssue(ctx, 2, "bad reason"); return e }, func() error { _, e := h.MergePull(ctx, 2, "bad", true); return e }, func() error { _, e := h.ReviewPull(ctx, 2, ReviewRequestChanges, ""); return e }}
	for i, f := range invalid {
		if e := f(); e == nil {
			t.Fatalf("case %d accepted", i)
		}
	}
	if len(*calls) != 0 {
		t.Fatalf("invalid input invoked gh: %v", *calls)
	}
	h.runner = func(context.Context, ...string) ([]byte, error) { return nil, errors.New("permission denied") }
	if _, e := h.CreatePull(ctx, CreatePullOptions{Title: "valid"}); e == nil {
		t.Fatal("mutation failure hidden")
	}
}
func TestGitHubCheckoutSafety(t *testing.T) {
	ctx := context.Background()
	g := testGit(t)
	testCommit(t, g, "a", "one", "initial")
	h, calls := fakeGitHub("checked out")
	h.Dir = g.Dir
	writeGitFile(t, g, "untracked", "do not lose")
	if _, e := h.CheckoutPull(ctx, 3); e == nil || len(*calls) != 0 {
		t.Fatal("dirty checkout attempted", e, *calls)
	}
	requireNoError(t, g.Discard(ctx, "untracked"))
	_, e := h.CheckoutPull(ctx, 3)
	requireNoError(t, e)
	if !reflect.DeepEqual((*calls)[0], []string{"pr", "checkout", "3"}) {
		t.Fatal(*calls)
	}
	testRun(t, g, "bisect", "start")
	if _, e := h.CheckoutPull(ctx, 3); e == nil || len(*calls) != 1 {
		t.Fatal("operation checkout attempted", e, *calls)
	}
}
