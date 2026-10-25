package native

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/charlzyx/orbit/internal/workspace"
)

type reportFakeGH struct {
	status workspace.GitHubStatus
	calls  int
	got    workspace.CreateIssueOptions
	err    error
}

func (f *reportFakeGH) Probe(context.Context) (workspace.GitHubStatus, error) { return f.status, nil }
func (f *reportFakeGH) CurrentUser(context.Context) (workspace.GitHubUser, error) {
	return workspace.GitHubUser{Login: "octocat"}, nil
}
func (f *reportFakeGH) CreateIssue(_ context.Context, o workspace.CreateIssueOptions) (string, error) {
	f.calls++
	f.got = o
	return "https://github.com/charlzyx/orbit/issues/1", f.err
}
func reportPNG(t *testing.T) []byte {
	t.Helper()
	im := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			im.Set(x, y, color.White)
		}
	}
	var b bytes.Buffer
	if e := png.Encode(&b, im); e != nil {
		t.Fatal(e)
	}
	return b.Bytes()
}
func TestReportPreparationConsentRedactionAndKinds(t *testing.T) {
	d := Diagnostics{OrbitVersion: "1.2.3", AppID: DefaultAppID, OS: "Linux", Arch: "x86_64", PiVersion: "1.4.0", GHUser: "octocat"}
	in := ReportInput{Kind: ReportBug, Title: "Crash & hang", What: "API_KEY=private-value at /home/name/project", Diagnostics: d, Home: "/home/name", RedactSensitive: true}
	p, e := PrepareReport(in)
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(p.Body, "private-value") || strings.Contains(p.Body, "/home/name") || strings.Contains(p.Body, "Diagnostics") {
		t.Fatal(p.Body)
	}
	if !strings.Contains(p.Body, "_No response._") {
		t.Fatal(p.Body)
	}
	u, _ := url.Parse(p.URL)
	if u.Query().Get("title") != in.Title || u.Query().Get("body") != p.Body || u.Query().Get("labels") != "bug" {
		t.Fatal(p.URL)
	}
	for _, k := range []ReportKind{ReportBug, ReportFeature, ReportOther} {
		in.Kind = k
		in.IncludeDiagnostics = true
		p, e = PrepareReport(in)
		if e != nil {
			t.Fatal(e)
		}
		heading, _ := k.Headings()
		if !strings.HasPrefix(p.Body, "## "+heading) || !strings.Contains(p.Body, "**Filed by:** @octocat") {
			t.Fatal(p.Body)
		}
	}
	fake := &reportFakeGH{status: workspace.GitHubStatus{Installed: true, Authenticated: true}}
	if _, e = FileReport(context.Background(), fake, p, false); e == nil || fake.calls != 0 {
		t.Fatal("publication without review")
	}
	in.Repository = "owner/name?redirect=elsewhere"
	if _, e = PrepareReport(in); e == nil {
		t.Fatal("invalid repo accepted")
	}
}
func TestScreenshotRedactionSanitizationAndCaps(t *testing.T) {
	shot, e := NewReportScreenshot(`../../private\shot one.jpg`, reportPNG(t))
	if e != nil {
		t.Fatal(e)
	}
	if shot.Name != "shot_one.png" {
		t.Fatal(shot.Name)
	}
	redacted, e := RedactScreenshot(shot, []image.Rectangle{image.Rect(1, 2, 5, 6)})
	if e != nil {
		t.Fatal(e)
	}
	im, _, _ := image.Decode(bytes.NewReader(redacted.PNG))
	r, g, b, a := im.At(2, 3).RGBA()
	if r != 0 || g != 0 || b != 0 || a != 65535 {
		t.Fatal("redaction not opaque")
	}
	original, _, _ := image.Decode(bytes.NewReader(shot.PNG))
	r, _, _, _ = original.At(2, 3).RGBA()
	if r == 0 {
		t.Fatal("original altered")
	}
	if _, e = NewReportScreenshot("secret.txt", []byte("secret")); e == nil {
		t.Fatal("non-image accepted")
	}
	_, e = PrepareReport(ReportInput{Kind: ReportBug, Title: "t", What: "d", Screenshots: make([]ReportScreenshot, 7)})
	if e == nil {
		t.Fatal("screenshot limit not enforced")
	}
}
func TestReportFilingFallbackAndNoDuplicateOnUncertainResult(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	shot, _ := NewReportScreenshot("a.png", reportPNG(t))
	p, e := PrepareReport(ReportInput{Kind: ReportBug, Title: "crash", What: "details", Screenshots: []ReportScreenshot{shot, shot}})
	if e != nil {
		t.Fatal(e)
	}
	f := &reportFakeGH{status: workspace.GitHubStatus{Installed: true, Authenticated: true}, err: errors.New("unknown flag: --attach")}
	result, e := FileReport(context.Background(), f, p, true)
	if e != nil || result.Created || result.ScreenshotDirectory == "" {
		t.Fatal(result, e)
	}
	if tempEntryActive(filepath.Base(result.ScreenshotDirectory)) {
		t.Fatal("finished report cannot be cleaned from Privacy")
	}
	entries, _ := os.ReadDir(result.ScreenshotDirectory)
	if len(entries) != 2 {
		t.Fatal(entries)
	}
	for _, entry := range entries {
		if info, e := os.Stat(filepath.Join(result.ScreenshotDirectory, entry.Name())); e != nil || info.Mode().Perm() != 0600 {
			t.Fatal("unsafe screenshot file", e)
		}
	}
	if !strings.Contains(result.URL, "Screenshots") {
		t.Fatal(result.URL)
	}
	f.err = errors.New("connection reset after upload")
	result, e = FileReport(context.Background(), f, p, true)
	if e == nil || result.URL != "" || f.calls != 2 {
		t.Fatal("ambiguous failure retried or hidden", result, e)
	}
	f.err = nil
	result, e = FileReport(context.Background(), f, p, true)
	if e != nil || !result.Created || len(f.got.Attachments) != 2 {
		t.Fatal(result, e)
	}
	if _, e = os.Stat(filepath.Dir(f.got.Attachments[0])); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("successful upload staging not removed")
	}
	if !tempEntryActive("bug-report-pid-" + fmtPID() + "-test") {
		t.Fatal("in-flight report not privacy guarded")
	}
}
func fmtPID() string { return strconv.Itoa(os.Getpid()) }
func TestDiagnosticsNeverPropagatesCLIErrorPayload(t *testing.T) {
	s := Diagnostics{OrbitVersion: "v1\nINJECT", PiVersion: "sk-abcdefghijklmnop"}.Markdown()
	if strings.Contains(s, "sk-") || strings.Contains(s, "v1\nINJECT") {
		t.Fatal(s)
	}
	about := AppAbout("dev", "bad/../../target")
	if about.RepositoryURL != "https://github.com/charlzyx/orbit" {
		t.Fatal(about)
	}
}
