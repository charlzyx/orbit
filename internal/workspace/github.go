package workspace

// GitHub operations are delegated to the official gh CLI. Credentials remain
// owned by gh; this package neither stores tokens nor calls GitHub directly.
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type GitHub struct {
	Dir, Binary, GitBinary string
	runner                 func(context.Context, ...string) ([]byte, error)
}

func (h GitHub) Run(ctx context.Context, args ...string) ([]byte, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if h.runner != nil {
		return h.runner(ctx, args...)
	}
	bin := h.Binary
	if bin == "" {
		bin = os.Getenv("GH_BIN")
	}
	if bin == "" {
		bin = "gh"
		if p, e := exec.LookPath(bin); e == nil {
			bin = p
		} else {
			home, _ := os.UserHomeDir()
			for _, dir := range []string{"/opt/homebrew/bin", "/usr/local/bin", "/usr/bin", filepath.Join(home, ".local", "bin"), filepath.Join(home, "scoop", "shims")} {
				p := filepath.Join(dir, "gh")
				if s, e := os.Stat(p); e == nil && !s.IsDir() {
					bin = p
					break
				}
			}
		}
	}
	return commandOutput(ctx, h.Dir, bin, []string{"GH_PROMPT_DISABLED=1", "GH_NO_UPDATE_NOTIFIER=1", "GH_PAGER=cat", "NO_COLOR=1", "CLICOLOR=0"}, args)
}
func (h GitHub) text(ctx context.Context, args ...string) (string, error) {
	b, e := h.Run(ctx, args...)
	return strings.TrimSpace(string(b)), e
}
func ghJSON[T any](ctx context.Context, h GitHub, args ...string) (T, error) {
	var v T
	b, e := h.Run(ctx, args...)
	if e != nil {
		return v, e
	}
	if e = json.Unmarshal(b, &v); e != nil {
		return v, fmt.Errorf("invalid GitHub CLI JSON: %w", e)
	}
	return v, nil
}

type GitHubStatus struct {
	Installed, Authenticated, Supported bool
	Version, Detail                     string
}

var ghVersionPattern = regexp.MustCompile(`(?:^|\s)v?([0-9]+)\.([0-9]+)\.([0-9]+)`)

func ParseGitHubVersion(raw string) ([3]int, bool) {
	var v [3]int
	m := ghVersionPattern.FindStringSubmatch(raw)
	if len(m) != 4 {
		return v, false
	}
	for i := range v {
		v[i], _ = strconv.Atoi(m[i+1])
	}
	return v, true
}
func (h GitHub) Probe(ctx context.Context) (GitHubStatus, error) {
	var status GitHubStatus
	v, e := h.text(ctx, "--version")
	if e != nil {
		if ctx.Err() != nil {
			return status, ctx.Err()
		}
		status.Detail = e.Error()
		return status, nil
	}
	status.Installed = true
	status.Version = strings.Split(v, "\n")[0]
	version, known := ParseGitHubVersion(status.Version)
	status.Supported = !known || version[0] > 2 || (version[0] == 2 && version[1] >= 20)
	_, e = h.Run(ctx, "auth", "status")
	if e != nil {
		if ctx.Err() != nil {
			return status, ctx.Err()
		}
		status.Detail = e.Error()
		return status, nil
	}
	status.Authenticated = true
	status.Detail = status.Version
	if !status.Supported {
		status.Detail += "; gh 2.20.0 or newer is required"
	}
	return status, nil
}

type GitHubUser struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}
type GitHubLabel struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}
type GitHubComment struct {
	Author    *GitHubUser `json:"author"`
	Body      string      `json:"body"`
	CreatedAt string      `json:"createdAt"`
	URL       string      `json:"url"`
}
type GitHubIssue struct {
	Number    int             `json:"number"`
	Title     string          `json:"title"`
	State     string          `json:"state"`
	Body      string          `json:"body"`
	Author    *GitHubUser     `json:"author"`
	Labels    []GitHubLabel   `json:"labels"`
	Assignees []GitHubUser    `json:"assignees"`
	Comments  []GitHubComment `json:"comments"`
	CreatedAt string          `json:"createdAt"`
	UpdatedAt string          `json:"updatedAt"`
	URL       string          `json:"url"`
}

func (i GitHubIssue) IsOpen() bool { return strings.EqualFold(i.State, "open") }

type IssueFilter struct {
	State, Label, Assignee, Search string
	Limit                          int
}

func (f IssueFilter) Active() bool {
	return (f.State != "" && f.State != "open") || f.Label != "" || f.Assignee != "" || f.Search != ""
}

type CreateIssueOptions struct {
	Repo, Title, Body              string
	Labels, Assignees, Attachments []string
}

const issueListFields = "number,title,state,author,labels,assignees,comments,createdAt,updatedAt,url"
const issueViewFields = "number,title,state,body,author,labels,assignees,comments,createdAt,updatedAt,url"

func (h GitHub) CurrentUser(ctx context.Context) (GitHubUser, error) {
	return ghJSON[GitHubUser](ctx, h, "api", "user")
}
func ghState(state string, pull bool) (string, error) {
	if state == "" {
		return "open", nil
	}
	switch state {
	case "open", "closed", "all":
		return state, nil
	case "merged":
		if pull {
			return state, nil
		}
	}
	return "", fmt.Errorf("invalid state filter %q", state)
}
func addFlag(args []string, flag, value string) []string {
	if strings.TrimSpace(value) != "" {
		return append(args, flag, value)
	}
	return args
}
func ghLimit(limit int) string {
	if limit <= 0 {
		limit = 50
	}
	return strconv.Itoa(limit)
}
func (h GitHub) Issues(ctx context.Context, f IssueFilter) ([]GitHubIssue, error) {
	state, e := ghState(f.State, false)
	if e != nil {
		return nil, e
	}
	args := []string{"issue", "list", "--json", issueListFields, "--limit", ghLimit(f.Limit), "--state", state}
	args = addFlag(args, "--label", f.Label)
	args = addFlag(args, "--assignee", f.Assignee)
	args = addFlag(args, "--search", f.Search)
	return ghJSON[[]GitHubIssue](ctx, h, args...)
}
func numberArg(n int) (string, error) {
	if n <= 0 {
		return "", errors.New("issue or pull request number must be positive")
	}
	return strconv.Itoa(n), nil
}
func (h GitHub) Issue(ctx context.Context, number int) (GitHubIssue, error) {
	n, e := numberArg(number)
	if e != nil {
		return GitHubIssue{}, e
	}
	return ghJSON[GitHubIssue](ctx, h, "issue", "view", n, "--json", issueViewFields)
}
func (h GitHub) CreateIssue(ctx context.Context, o CreateIssueOptions) (string, error) {
	if strings.TrimSpace(o.Title) == "" {
		return "", errors.New("issue title is required")
	}
	args := addFlag([]string{"issue", "create"}, "--repo", o.Repo)
	args = append(args, "--title", o.Title, "--body", o.Body)
	for _, s := range o.Attachments {
		args = append(args, "--attach", s)
	}
	for _, s := range o.Labels {
		args = append(args, "--label", s)
	}
	for _, s := range o.Assignees {
		args = append(args, "--assignee", s)
	}
	return h.text(ctx, args...)
}
func (h GitHub) CommentIssue(ctx context.Context, number int, body string) (string, error) {
	return h.comment(ctx, "issue", number, body)
}
func (h GitHub) comment(ctx context.Context, kind string, number int, body string) (string, error) {
	n, e := numberArg(number)
	if e != nil {
		return "", e
	}
	if strings.TrimSpace(body) == "" {
		return "", errors.New("comment cannot be empty")
	}
	return h.text(ctx, kind, "comment", n, "--body", body)
}
func (h GitHub) EditIssueLabels(ctx context.Context, number int, add, remove []string) (string, error) {
	n, e := numberArg(number)
	if e != nil {
		return "", e
	}
	if len(add)+len(remove) == 0 {
		return "", nil
	}
	args := []string{"issue", "edit", n}
	for _, s := range add {
		args = append(args, "--add-label", s)
	}
	for _, s := range remove {
		args = append(args, "--remove-label", s)
	}
	return h.text(ctx, args...)
}
func (h GitHub) CloseIssue(ctx context.Context, number int, reason string) (string, error) {
	n, e := numberArg(number)
	if e != nil {
		return "", e
	}
	if reason != "" && reason != "completed" && reason != "not planned" {
		return "", errors.New("invalid issue close reason")
	}
	return h.text(ctx, addFlag([]string{"issue", "close", n}, "--reason", reason)...)
}
func (h GitHub) ReopenIssue(ctx context.Context, number int) (string, error) {
	return h.numberAction(ctx, "issue", "reopen", number)
}
func (h GitHub) numberAction(ctx context.Context, kind, action string, number int) (string, error) {
	n, e := numberArg(number)
	if e != nil {
		return "", e
	}
	return h.text(ctx, kind, action, n)
}
func (h GitHub) Labels(ctx context.Context) ([]GitHubLabel, error) {
	return ghJSON[[]GitHubLabel](ctx, h, "label", "list", "--json", "name,color,description", "--limit", "100")
}

type CheckBucket string

const (
	CheckPassed  CheckBucket = "passed"
	CheckFailed  CheckBucket = "failed"
	CheckPending CheckBucket = "pending"
	CheckSkipped CheckBucket = "skipped"
	CheckUnknown CheckBucket = "unknown"
)

type GitHubCheck struct {
	Name        string `json:"name"`
	Context     string `json:"context"`
	State       string `json:"state"`
	Status      string `json:"status"`
	Conclusion  string `json:"conclusion"`
	Bucket      string `json:"bucket"`
	Description string `json:"description"`
	DetailsURL  string `json:"detailsUrl"`
	TargetURL   string `json:"targetUrl"`
	Link        string `json:"link"`
}

func (c GitHubCheck) BucketKind() CheckBucket {
	if c.Bucket != "" {
		switch c.Bucket {
		case "pass":
			return CheckPassed
		case "fail":
			return CheckFailed
		case "pending":
			return CheckPending
		case "skipping", "cancel":
			return CheckSkipped
		default:
			return CheckUnknown
		}
	}
	raw := c.Conclusion
	if raw == "" {
		raw = c.State
	}
	if raw == "" {
		raw = c.Status
	}
	switch strings.ToUpper(raw) {
	case "SUCCESS":
		return CheckPassed
	case "FAILURE", "ERROR", "TIMED_OUT", "STARTUP_FAILURE", "ACTION_REQUIRED":
		return CheckFailed
	case "PENDING", "QUEUED", "IN_PROGRESS", "WAITING", "REQUESTED":
		return CheckPending
	case "SKIPPED", "NEUTRAL", "STALE", "CANCELLED":
		return CheckSkipped
	}
	return CheckUnknown
}
func (c GitHubCheck) Label() string {
	if c.Name != "" {
		return c.Name
	}
	return c.Context
}
func (c GitHubCheck) URL() string {
	if c.DetailsURL != "" {
		return c.DetailsURL
	}
	if c.TargetURL != "" {
		return c.TargetURL
	}
	return c.Link
}

type GitHubReview struct {
	Author      *GitHubUser `json:"author"`
	Body        string      `json:"body"`
	State       string      `json:"state"`
	SubmittedAt string      `json:"submittedAt"`
	URL         string      `json:"url"`
}
type GitHubPullCommit struct {
	Headline string       `json:"messageHeadline"`
	OID      string       `json:"oid"`
	Authors  []GitHubUser `json:"authors"`
}
type GitHubPullFile struct {
	Path       string `json:"path"`
	Additions  int    `json:"additions"`
	Deletions  int    `json:"deletions"`
	ChangeType string `json:"changeType"`
}
type GitHubPull struct {
	Number           int                `json:"number"`
	Title            string             `json:"title"`
	State            string             `json:"state"`
	IsDraft          bool               `json:"isDraft"`
	Body             string             `json:"body"`
	Author           *GitHubUser        `json:"author"`
	Labels           []GitHubLabel      `json:"labels"`
	ReviewDecision   string             `json:"reviewDecision"`
	Mergeable        string             `json:"mergeable"`
	MergeStateStatus string             `json:"mergeStateStatus"`
	HeadRef          string             `json:"headRefName"`
	BaseRef          string             `json:"baseRefName"`
	Checks           []GitHubCheck      `json:"statusCheckRollup"`
	Additions        int                `json:"additions"`
	Deletions        int                `json:"deletions"`
	ChangedFiles     int                `json:"changedFiles"`
	CreatedAt        string             `json:"createdAt"`
	UpdatedAt        string             `json:"updatedAt"`
	URL              string             `json:"url"`
	Comments         []GitHubComment    `json:"comments"`
	Reviews          []GitHubReview     `json:"reviews"`
	Commits          []GitHubPullCommit `json:"commits"`
	Files            []GitHubPullFile   `json:"files"`
}

func (p GitHubPull) IsOpen() bool   { return strings.EqualFold(p.State, "open") }
func (p GitHubPull) IsMerged() bool { return strings.EqualFold(p.State, "merged") }
func (p GitHubPull) ChecksSummary() (CheckBucket, bool) {
	if len(p.Checks) == 0 {
		return "", false
	}
	pending, passed := false, false
	for _, c := range p.Checks {
		switch c.BucketKind() {
		case CheckFailed:
			return CheckFailed, true
		case CheckPending:
			pending = true
		case CheckPassed:
			passed = true
		}
	}
	if pending {
		return CheckPending, true
	}
	if passed {
		return CheckPassed, true
	}
	return CheckSkipped, true
}

type PullFilter struct {
	State, Base, Author, Search string
	Limit                       int
}

func (f PullFilter) Active() bool {
	return (f.State != "" && f.State != "open") || f.Base != "" || f.Author != "" || f.Search != ""
}

type CreatePullOptions struct {
	Title, Body, Base string
	Draft             bool
	Labels            []string
}
type PullMergeMethod string

const (
	PullMerge  PullMergeMethod = "merge"
	PullSquash PullMergeMethod = "squash"
	PullRebase PullMergeMethod = "rebase"
)

type PullReviewAction string

const (
	ReviewApprove        PullReviewAction = "approve"
	ReviewRequestChanges PullReviewAction = "request-changes"
	ReviewComment        PullReviewAction = "comment"
)
const pullListFields = "number,title,state,isDraft,author,labels,reviewDecision,mergeable,mergeStateStatus,headRefName,baseRefName,statusCheckRollup,additions,deletions,changedFiles,updatedAt,url"
const pullViewFields = "number,title,state,isDraft,body,author,labels,reviewDecision,mergeable,mergeStateStatus,headRefName,baseRefName,statusCheckRollup,additions,deletions,changedFiles,createdAt,updatedAt,url,comments,reviews,commits,files"

func (h GitHub) Pulls(ctx context.Context, f PullFilter) ([]GitHubPull, error) {
	state, e := ghState(f.State, true)
	if e != nil {
		return nil, e
	}
	args := []string{"pr", "list", "--json", pullListFields, "--limit", ghLimit(f.Limit), "--state", state}
	args = addFlag(args, "--base", f.Base)
	args = addFlag(args, "--author", f.Author)
	args = addFlag(args, "--search", f.Search)
	return ghJSON[[]GitHubPull](ctx, h, args...)
}
func (h GitHub) Pull(ctx context.Context, number int) (GitHubPull, error) {
	n, e := numberArg(number)
	if e != nil {
		return GitHubPull{}, e
	}
	return ghJSON[GitHubPull](ctx, h, "pr", "view", n, "--json", pullViewFields)
}
func (h GitHub) CreatePull(ctx context.Context, o CreatePullOptions) (string, error) {
	if strings.TrimSpace(o.Title) == "" {
		return "", errors.New("pull request title is required")
	}
	args := []string{"pr", "create", "--title", o.Title, "--body", o.Body}
	args = addFlag(args, "--base", o.Base)
	if o.Draft {
		args = append(args, "--draft")
	}
	for _, s := range o.Labels {
		args = append(args, "--label", s)
	}
	return h.text(ctx, args...)
}
func (h GitHub) CommentPull(ctx context.Context, number int, body string) (string, error) {
	return h.comment(ctx, "pr", number, body)
}
func (h GitHub) ReviewPull(ctx context.Context, number int, action PullReviewAction, body string) (string, error) {
	n, e := numberArg(number)
	if e != nil {
		return "", e
	}
	switch action {
	case ReviewApprove, ReviewRequestChanges, ReviewComment:
	default:
		return "", errors.New("invalid pull request review action")
	}
	if (action == ReviewRequestChanges || action == ReviewComment) && strings.TrimSpace(body) == "" {
		return "", errors.New("review body is required")
	}
	return h.text(ctx, addFlag([]string{"pr", "review", n, "--" + string(action)}, "--body", body)...)
}
func (h GitHub) MergePull(ctx context.Context, number int, method PullMergeMethod, deleteBranch bool) (string, error) {
	n, e := numberArg(number)
	if e != nil {
		return "", e
	}
	if method == "" {
		method = PullMerge
	}
	switch method {
	case PullMerge, PullSquash, PullRebase:
	default:
		return "", errors.New("invalid pull request merge method")
	}
	args := []string{"pr", "merge", n, "--" + string(method)}
	if deleteBranch {
		args = append(args, "--delete-branch")
	}
	return h.text(ctx, args...)
}
func (h GitHub) ClosePull(ctx context.Context, number int) (string, error) {
	return h.numberAction(ctx, "pr", "close", number)
}
func (h GitHub) ReopenPull(ctx context.Context, number int) (string, error) {
	return h.numberAction(ctx, "pr", "reopen", number)
}
func (h GitHub) CheckoutPull(ctx context.Context, number int) (string, error) {
	if _, e := numberArg(number); e != nil {
		return "", e
	}
	g := Git{Dir: h.Dir, Binary: h.GitBinary}
	op, e := g.Operation(ctx)
	if e != nil {
		return "", e
	}
	if op != OperationNone {
		return "", fmt.Errorf("cannot check out pull request during %s", op)
	}
	rows, e := g.Status(ctx)
	if e != nil {
		return "", e
	}
	if len(rows) != 0 {
		return "", errors.New("commit or stash local changes before checking out a pull request")
	}
	return h.numberAction(ctx, "pr", "checkout", number)
}
