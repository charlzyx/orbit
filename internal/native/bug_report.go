package native

// Report preparation is entirely local. The only publishing operation is
// FileReport, which requires explicit confirmation of an already reviewed
// PreparedReport. No environment, log, credential or session files are read.
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/charlzyx/orbit/internal/workspace"
	_ "golang.org/x/image/webp"
)

const DefaultReportRepository = "charlzyx/orbit"
const DefaultAppID = "dev.orbit.pi"
const MaxReportScreenshots = 6
const MaxReportImageBytes = 16 << 20
const MaxReportImagePixels = 20_000_000

type ReportKind string

const (
	ReportBug     ReportKind = "bug"
	ReportFeature ReportKind = "feature"
	ReportOther   ReportKind = "other"
)

func (k ReportKind) Label() string {
	switch k {
	case ReportBug:
		return "bug"
	case ReportFeature:
		return "enhancement"
	}
	return ""
}
func (k ReportKind) Headings() (string, string) {
	switch k {
	case ReportFeature:
		return "Problem", "Proposed solution"
	case ReportOther:
		return "Details", "Additional details"
	}
	return "What happened?", "Steps to reproduce"
}
func (k ReportKind) valid() bool { return k == ReportBug || k == ReportFeature || k == ReportOther }

type ReportGitHub interface {
	Probe(context.Context) (workspace.GitHubStatus, error)
	CurrentUser(context.Context) (workspace.GitHubUser, error)
	CreateIssue(context.Context, workspace.CreateIssueOptions) (string, error)
}
type DiagnosticOptions struct {
	Version, AppID, PiBinary string
	IncludeGitHubIdentity    bool
	GitHub                   ReportGitHub
}
type Diagnostics struct{ OrbitVersion, AppID, OS, Arch, Install, PiVersion, GHVersion, GHUser string }

func CollectDiagnostics(ctx context.Context, o DiagnosticOptions) Diagnostics {
	host := ProbeHost(ctx)
	d := Diagnostics{OrbitVersion: o.Version, AppID: o.AppID, OS: host.Label, Arch: host.Arch, Install: "Source build"}
	if d.AppID == "" {
		d.AppID = DefaultAppID
	}
	if exe, e := os.Executable(); e == nil && strings.Contains(filepath.ToSlash(exe), ".app/Contents/MacOS/") {
		d.Install = "App bundle"
	}
	binary := o.PiBinary
	if binary == "" {
		binary = LocateExecutable("pi")
	}
	if binary != "" {
		d.PiVersion, _ = ExecutableVersion(ctx, binary)
	}
	if gh := LocateExecutable("gh"); gh != "" {
		d.GHVersion, _ = ExecutableVersion(ctx, gh)
	}
	// Account identity is a separate opt-in. Merely opening Settings does not
	// contact GitHub or inspect its credentials.
	if o.IncludeGitHubIdentity && o.GitHub != nil {
		if status, e := o.GitHub.Probe(ctx); e == nil && status.Authenticated {
			if user, e := o.GitHub.CurrentUser(ctx); e == nil {
				d.GHUser = user.Login
			}
		}
	}
	return d
}
func diagnosticValue(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, s)
	s = RedactReportText(s, "", nil)
	if len([]rune(s)) > 200 {
		s = string([]rune(s)[:200]) + "…"
	}
	return strings.NewReplacer("<", "&lt;", ">", "&gt;", "*", "\\*", "`", "\\`").Replace(s)
}
func (d Diagnostics) Markdown() string {
	rows := [][2]string{{"Orbit", d.OrbitVersion}, {"App ID", d.AppID}, {"OS", d.OS}, {"Architecture", d.Arch}, {"Install", d.Install}, {"pi CLI", d.PiVersion}, {"gh CLI", d.GHVersion}}
	if d.GHUser != "" {
		rows = append(rows, [2]string{"Filed by", "@" + d.GHUser})
	}
	var b strings.Builder
	b.WriteString("<details>\n<summary>Diagnostics</summary>\n\n")
	for _, r := range rows {
		if r[1] != "" {
			fmt.Fprintf(&b, "- **%s:** %s\n", r[0], diagnosticValue(r[1]))
		}
	}
	b.WriteString("\n</details>\n")
	return b.String()
}

var reportSecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:bearer\s+)[A-Za-z0-9._~+/=-]{8,}`),
	regexp.MustCompile(`(?:gh[pousr]_[A-Za-z0-9_]{12,}|github_pat_[A-Za-z0-9_]{12,}|sk-[A-Za-z0-9_-]{12,}|AIza[A-Za-z0-9_-]{20,})`),
	regexp.MustCompile(`(?i)(?:api[_-]?key|access[_-]?token|refresh[_-]?token|password|client[_-]?secret)\s*[:=]\s*["']?[^\s"',;]+`),
	regexp.MustCompile(`https?://[^\s/:@]+:[^\s/@]+@`),
}

// RedactReportText is a review aid, not a promise to identify every secret.
// Known values must be explicitly provided; this function never loads auth.
func RedactReportText(s, home string, known []string) string {
	for _, secret := range known {
		if secret != "" {
			s = strings.ReplaceAll(s, secret, "[REDACTED]")
		}
	}
	for _, p := range reportSecretPatterns {
		s = p.ReplaceAllString(s, "[REDACTED]")
	}
	if home != "" {
		s = strings.ReplaceAll(s, home, "~")
		s = strings.ReplaceAll(s, filepath.ToSlash(home), "~")
	}
	return s
}

type ReportScreenshot struct {
	Name          string
	PNG           []byte
	Width, Height int
}

// NewReportScreenshot decodes and re-encodes a chosen image, dropping metadata
// (including EXIF GPS). Animated images use the reviewed first frame only.
func NewReportScreenshot(name string, data []byte) (ReportScreenshot, error) {
	if len(data) == 0 || len(data) > MaxReportImageBytes {
		return ReportScreenshot{}, errors.New("screenshot must be between 1 byte and 16 MiB")
	}
	cfg, _, e := image.DecodeConfig(bytes.NewReader(data))
	if e != nil {
		return ReportScreenshot{}, fmt.Errorf("decode screenshot: %w", e)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > MaxReportImagePixels {
		return ReportScreenshot{}, errors.New("screenshot exceeds 20 megapixels")
	}
	img, _, e := image.Decode(bytes.NewReader(data))
	if e != nil {
		return ReportScreenshot{}, e
	}
	var b bytes.Buffer
	if e = png.Encode(&b, img); e != nil {
		return ReportScreenshot{}, e
	}
	if b.Len() > MaxReportImageBytes {
		return ReportScreenshot{}, errors.New("normalized screenshot exceeds 16 MiB")
	}
	name = strings.TrimSuffix(sanitizeScreenshotName(name), filepath.Ext(sanitizeScreenshotName(name))) + ".png"
	return ReportScreenshot{Name: name, PNG: b.Bytes(), Width: cfg.Width, Height: cfg.Height}, nil
}
func LoadReportScreenshot(path string) (ReportScreenshot, error) {
	f, e := os.Open(path)
	if e != nil {
		return ReportScreenshot{}, e
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil {
		return ReportScreenshot{}, e
	}
	if !info.Mode().IsRegular() {
		return ReportScreenshot{}, errors.New("screenshot is not a regular file")
	}
	b, e := io.ReadAll(io.LimitReader(f, MaxReportImageBytes+1))
	if e != nil {
		return ReportScreenshot{}, e
	}
	return NewReportScreenshot(filepath.Base(path), b)
}

// RedactScreenshot irreversibly paints opaque boxes into a copy. The original
// chosen file is never modified, nor retained in the returned screenshot.
func RedactScreenshot(s ReportScreenshot, rects []image.Rectangle) (ReportScreenshot, error) {
	source, _, e := image.Decode(bytes.NewReader(s.PNG))
	if e != nil {
		return ReportScreenshot{}, e
	}
	dst := image.NewNRGBA(source.Bounds())
	draw.Draw(dst, dst.Bounds(), source, source.Bounds().Min, draw.Src)
	for _, r := range rects {
		draw.Draw(dst, r.Intersect(dst.Bounds()), image.NewUniform(color.Black), image.Point{}, draw.Src)
	}
	var b bytes.Buffer
	if e = png.Encode(&b, dst); e != nil {
		return ReportScreenshot{}, e
	}
	s.PNG = b.Bytes()
	return s, nil
}
func sanitizeScreenshotName(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if r > 127 || !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '-' || r == '_') {
			return '_'
		}
		return r
	}, name)
	name = strings.Trim(name, ".")
	if name == "" {
		name = "screenshot.png"
	}
	if len(name) > 120 {
		name = name[:120]
	}
	return name
}

type ReportInput struct {
	Repository         string
	Kind               ReportKind
	Title, What, Steps string
	IncludeDiagnostics bool
	Diagnostics        Diagnostics
	Screenshots        []ReportScreenshot
	RedactSensitive    bool
	Home               string
	KnownSecrets       []string
}
type PreparedReport struct {
	Repository                    string
	Kind                          ReportKind
	Title, Body, BrowserBody, URL string
	Screenshots                   []ReportScreenshot
}

func PrepareReport(in ReportInput) (PreparedReport, error) {
	p := PreparedReport{Repository: in.Repository, Kind: in.Kind, Title: strings.TrimSpace(in.Title)}
	if p.Repository == "" {
		p.Repository = DefaultReportRepository
	}
	if !validReportRepository(p.Repository) {
		return p, errors.New("report repository must be owner/name")
	}
	if !p.Kind.valid() {
		return p, errors.New("invalid report kind")
	}
	if p.Title == "" || strings.TrimSpace(in.What) == "" {
		return p, errors.New("a title and description are required")
	}
	if len(in.Screenshots) > MaxReportScreenshots {
		return p, errors.New("a report can include at most 6 screenshots")
	}
	if in.RedactSensitive {
		p.Title = RedactReportText(p.Title, in.Home, in.KnownSecrets)
		in.What = RedactReportText(in.What, in.Home, in.KnownSecrets)
		in.Steps = RedactReportText(in.Steps, in.Home, in.KnownSecrets)
	}
	if len(p.Title) > 1024 || len(in.What)+len(in.Steps) > 60*1024 {
		return p, errors.New("report text is too long")
	}
	what, steps := p.Kind.Headings()
	if strings.TrimSpace(in.Steps) == "" {
		in.Steps = "_No response._"
	}
	base := fmt.Sprintf("## %s\n\n%s\n\n## %s\n\n%s\n", what, in.What, steps, strings.TrimSpace(in.Steps))
	diagnostics := ""
	if in.IncludeDiagnostics {
		diagnostics = "\n---\n\n" + in.Diagnostics.Markdown()
	}
	p.Body = base + diagnostics
	p.BrowserBody = p.Body
	names := map[string]int{}
	for _, s := range in.Screenshots {
		checked, e := NewReportScreenshot(s.Name, s.PNG)
		if e != nil {
			return p, e
		}
		name := checked.Name
		names[name]++
		if names[name] > 1 {
			checked.Name = strings.TrimSuffix(name, ".png") + fmt.Sprintf("-%d.png", names[name])
		}
		p.Screenshots = append(p.Screenshots, checked)
	}
	if len(p.Screenshots) > 0 {
		p.BrowserBody = base + "\n## Screenshots\n\n<!-- Drag the screenshots here to attach them. -->\n" + diagnostics
	}
	p.URL = NewReportURL(p.Repository, p.Title, p.BrowserBody, p.Kind.Label())
	return p, nil
}
func validReportRepository(repo string) bool {
	return regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*/[A-Za-z0-9][A-Za-z0-9_.-]*$`).MatchString(repo)
}
func NewReportURL(repo, title, body, label string) string {
	if !validReportRepository(repo) {
		return ""
	}
	q := url.Values{"title": {title}, "body": {body}}
	if label != "" {
		q.Set("labels", label)
	}
	return "https://github.com/" + repo + "/issues/new?" + q.Encode()
}

type ReportFiled struct {
	Created                  bool
	URL, ScreenshotDirectory string
}

// FileReport does not open a browser, and never retries ambiguous gh failures.
// confirmed must come from a user action after reviewing this exact snapshot.
func FileReport(ctx context.Context, h ReportGitHub, p PreparedReport, confirmed bool) (result ReportFiled, err error) {
	if !confirmed {
		return ReportFiled{}, errors.New("review and confirm this report before sending")
	}
	if !validReportRepository(p.Repository) || p.Title == "" || p.Body == "" || !p.Kind.valid() {
		return ReportFiled{}, errors.New("invalid prepared report")
	}
	if len(p.Screenshots) > MaxReportScreenshots {
		return ReportFiled{}, errors.New("too many screenshots")
	}
	if e := ctx.Err(); e != nil {
		return ReportFiled{}, e
	}
	var paths []string
	directory := ""
	// Only in-flight staging is PID guarded. Once the worker finishes, browser
	// attachments and failed-report files become ordinary Privacy-cleanable data.
	defer func() {
		if directory == "" {
			return
		}
		prefix := fmt.Sprintf("bug-report-pid-%d-", os.Getpid())
		ready := filepath.Join(filepath.Dir(directory), "bug-report-"+strings.TrimPrefix(filepath.Base(directory), prefix))
		if _, e := os.Lstat(ready); !errors.Is(e, os.ErrNotExist) {
			return
		}
		if e := os.Rename(directory, ready); e == nil && result.ScreenshotDirectory == directory {
			result.ScreenshotDirectory = ready
		}
	}()

	if len(p.Screenshots) > 0 {
		root, e := EnsureTempDir()
		if e != nil {
			return ReportFiled{}, e
		}
		directory, e = os.MkdirTemp(root, fmt.Sprintf("bug-report-pid-%d-", os.Getpid()))
		if e != nil {
			return ReportFiled{}, e
		}
		for i, s := range p.Screenshots {
			path := filepath.Join(directory, fmt.Sprintf("%d-%s", i+1, sanitizeScreenshotName(s.Name)))
			if e = os.WriteFile(path, s.PNG, 0600); e != nil {
				os.RemoveAll(directory)
				return ReportFiled{}, e
			}
			paths = append(paths, path)
		}
	}
	status := workspace.GitHubStatus{}
	if h != nil {
		var e error
		status, e = h.Probe(ctx)
		if e != nil {
			return ReportFiled{ScreenshotDirectory: directory}, e
		}
	}
	if status.Installed && status.Authenticated {
		labels := []string{}
		if label := p.Kind.Label(); label != "" {
			labels = append(labels, label)
		}
		location, e := h.CreateIssue(ctx, workspace.CreateIssueOptions{Repo: p.Repository, Title: p.Title, Body: p.Body, Labels: labels, Attachments: paths})
		if e == nil {
			if directory != "" {
				_ = os.RemoveAll(directory)
			}
			return ReportFiled{Created: true, URL: strings.TrimSpace(location)}, nil
		}
		// The unsupported flag is parsed before any network mutation. Other errors
		// may follow a successful creation and MUST NOT trigger a duplicate issue.
		msg := strings.ToLower(e.Error())
		unsupported := len(paths) > 0 && (strings.Contains(msg, "unknown flag: --attach") || strings.Contains(msg, "unknown flag: attach") || strings.Contains(msg, "unknown shorthand flag"))
		if !unsupported {
			return ReportFiled{ScreenshotDirectory: directory}, fmt.Errorf("GitHub submission failed; check for an existing issue before retrying: %w", e)
		}
	}
	return ReportFiled{URL: NewReportURL(p.Repository, p.Title, p.BrowserBody, p.Kind.Label()), ScreenshotDirectory: directory}, nil
}

type AboutInfo struct{ Name, Version, AppID, Toolkit, ToolkitVersion, Runtime, License, RepositoryURL, IssuesURL, ReleasesURL string }

func AppAbout(version, repository string) AboutInfo {
	if !validReportRepository(repository) {
		repository = DefaultReportRepository
	}
	root := "https://github.com/" + repository
	return AboutInfo{Name: "Orbit Pi", Version: version, AppID: DefaultAppID, Toolkit: "MyGo native UI", ToolkitVersion: "0.6.4-dev (9c69cf829ece)", Runtime: "Pi CLI over direct stdio RPC", License: "Apache-2.0", RepositoryURL: root, IssuesURL: root + "/issues", ReleasesURL: root + "/releases"}
}
