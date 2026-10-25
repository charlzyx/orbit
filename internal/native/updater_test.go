package native

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseAppcastAndVersionHistory(t *testing.T) {
	raw := []byte(`<rss xmlns:sparkle="http://www.andymatuschak.org/xml-namespaces/sparkle"><channel><item><sparkle:shortVersionString>0.9.0</sparkle:shortVersionString><description>Faster &amp; safer</description><enclosure url="https://release.test/a" length="10" sparkle:edSignature="old"/></item><item><enclosure url="https://release.test/b" length="20" sparkle:shortVersionString="0.10.0" sparkle:edSignature="new"/></item><item><sparkle:shortVersionString>9.0</sparkle:shortVersionString><enclosure url="https://release.test/unsigned" /></item></channel></rss>`)
	items, err := ParseAppcast(raw)
	if err != nil || len(items) != 2 || items[0].Version != "0.10.0" || items[1].Notes != "Faster & safer" {
		t.Fatal(items, err)
	}
	for _, tc := range []struct {
		a, b string
		want int
	}{{"1.2", "1.2.0", 0}, {"1.2.1", "1.2", 1}, {"0.1.9", "0.1.10", -1}, {"1.2.0+build.7", "1.2.0", 0}, {"-", "", 0}} {
		if got := CompareVersions(tc.a, tc.b); got != tc.want {
			t.Errorf("%s %s: %d", tc.a, tc.b, got)
		}
	}
}
func TestVerifyArtifactRejectsTamperingAndLength(t *testing.T) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("signed disposable artifact")
	r := Release{Length: int64(len(data)), Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, data))}
	file := filepath.Join(t.TempDir(), "artifact")
	os.WriteFile(file, data, 0600)
	if err = VerifyArtifact(file, r, public); err != nil {
		t.Fatal(err)
	}
	data[0] ^= 1
	os.WriteFile(file, data, 0600)
	if VerifyArtifact(file, r, public) == nil {
		t.Fatal("accepted tampered artifact")
	}
	r.Length++
	if VerifyArtifact(file, r, public) == nil {
		t.Fatal("accepted wrong length")
	}
}

type archiveEntry struct {
	name, body string
	kind       byte
	mode       int64
}

func archive(t *testing.T, entries []archiveEntry) []byte {
	t.Helper()
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		kind := e.kind
		if kind == 0 {
			kind = tar.TypeReg
		}
		mode := e.mode
		if mode == 0 {
			mode = 0644
		}
		h := &tar.Header{Name: e.name, Typeflag: kind, Mode: mode, Size: int64(len(e.body))}
		if kind != tar.TypeReg {
			h.Size = 0
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if kind == tar.TypeReg {
			if _, err := io.WriteString(tw, e.body); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	gz.Close()
	return b.Bytes()
}
func TestArchiveRejectsTraversalLinksDuplicatesAndWrongBuild(t *testing.T) {
	for name, entries := range map[string][]archiveEntry{"traversal": {{name: "release/../escape", body: "bad"}}, "absolute": {{name: "/release/file", body: "bad"}}, "symlink": {{name: "release/link", kind: tar.TypeSymlink}}, "duplicate": {{name: "release/file", body: "one"}, {name: "release/file", body: "two"}}, "wrong-platform": {{name: "other/file", body: "bad"}}, "backslash": {{name: `release/..\escape`, body: "bad"}}} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			file := filepath.Join(dir, "payload.tar.gz")
			os.WriteFile(file, archive(t, entries), 0600)
			if err := ExtractReleaseArchive(file, filepath.Join(dir, "out"), "release"); err == nil {
				t.Fatal("unsafe archive accepted")
			}
		})
	}
}

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestSignedUpdateEndToEndDisposableStaging(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix tar staging")
	}
	public, private, _ := ed25519.GenerateKey(rand.Reader)
	root := filepath.Join(t.TempDir(), "Orbit")
	os.MkdirAll(root, 0700)
	payload := archive(t, []archiveEntry{{name: "release", kind: tar.TypeDir}, {name: "release/bin", kind: tar.TypeDir}, {name: "release/bin/orbit", body: "#!/bin/sh\nexit 0\n", mode: 0755}, {name: "release/resources/icon", body: "icon"}})
	sig := base64.StdEncoding.EncodeToString(ed25519.Sign(private, payload))
	feed := fmt.Sprintf(`<rss xmlns:sparkle="http://www.andymatuschak.org/xml-namespaces/sparkle"><channel><item><sparkle:shortVersionString>2.0.0</sparkle:shortVersionString><description>Release notes</description><enclosure url="https://test.invalid/payload" length="%d" sparkle:edSignature="%s"/></item></channel></rss>`, len(payload), sig)
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		data := []byte(feed)
		if r.URL.Path == "/payload" {
			data = payload
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(data)), ContentLength: int64(len(data)), Header: http.Header{}, Request: r}, nil
	})}
	u := NewUpdater(UpdaterOptions{Version: "1.0.0", FeedURL: "https://test.invalid/feed", PublicKey: base64.StdEncoding.EncodeToString(public), Client: client, Layout: &InstallLayout{Root: root, Executable: "bin/orbit"}, ExpectedRoot: func(string) string { return "release" }})
	defer u.Close()
	if err := u.Check(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s := u.Snapshot(); s.Status != UpdateAvailable || s.Available.Version != "2.0.0" {
		t.Fatal(s)
	}
	if err := u.Stage(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s := u.Snapshot(); s.Status != UpdateStaged || s.Downloaded != int64(len(payload)) {
		t.Fatal(s)
	}
	if err := validateStaging(u.opts.Layout, u.staged.Directory); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(u.staged.Directory, "bin/orbit")); err != nil || !strings.HasPrefix(string(data), "#!/bin/sh") {
		t.Fatal(string(data), err)
	}
	if _, err := os.Stat(filepath.Join(root, "bin/orbit")); !os.IsNotExist(err) {
		t.Fatal("stage changed current install")
	}
	staged := u.staged.Directory
	if err := u.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(staged); !os.IsNotExist(err) {
		t.Fatal("stage not cleaned")
	}
}
func TestUpdateDisabledAndMalformedHelper(t *testing.T) {
	u := NewUpdater(UpdaterOptions{Version: "1"})
	if u.Snapshot().Enabled || u.Check(context.Background()) != ErrUpdatesDisabled {
		t.Fatal("unsigned dev build enabled")
	}
	for _, raw := range []string{"http://release.test/a", "file:///a", "https://user:secret@release.test/a", "https://release.test/a#fragment"} {
		if validDownloadURL(raw) == nil {
			t.Fatal(raw)
		}
	}
	if handled, code := RunInstallHelper([]string{updateHelperFlag, "--unexpected", "x"}); !handled || code == 0 {
		t.Fatal(handled, code)
	}
}

func TestClosingUpdaterCancelsPendingRead(t *testing.T) {
	public, _, _ := ed25519.GenerateKey(rand.Reader)
	entered := make(chan struct{})
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		close(entered)
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	u := NewUpdater(UpdaterOptions{FeedURL: "https://test.invalid/feed", PublicKey: base64.StdEncoding.EncodeToString(public), Client: client})
	done := make(chan error, 1)
	go func() { done <- u.Check(context.Background()) }()
	<-entered
	if err := u.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err == nil {
		t.Fatal("pending read survived closing updater")
	}
	if err := u.Check(context.Background()); err != os.ErrClosed {
		t.Fatal(err)
	}
}
