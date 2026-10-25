package native

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ReleaseFeedURL and ReleasePublicKey are supplied by a signed release build
// (-ldflags -X), never fetched from a feed or inherited from an upstream app.
// A development build with no key cannot replace itself with another product.
var ReleaseFeedURL, ReleasePublicKey string

const MaxFeedBytes int64 = 1 << 20
const MaxArtifactBytes int64 = 512 << 20
const MaxUnpackedBytes int64 = 1 << 30
const MaxArchiveEntries = 100000

var ErrUpdatesDisabled = errors.New("signed updates are not configured for this build")

type Release struct {
	Version, URL, Signature, Notes string
	Length                         int64
}
type UpdateStatus string

const (
	UpdateIdle        UpdateStatus = "idle"
	UpdateChecking    UpdateStatus = "checking"
	UpdateAvailable   UpdateStatus = "available"
	UpdateDownloading UpdateStatus = "downloading"
	UpdateStaged      UpdateStatus = "staged"
	UpdateInstalling  UpdateStatus = "installing"
	UpdateFailed      UpdateStatus = "failed"
)

type UpdateSnapshot struct {
	Enabled, CanInstall bool
	Status              UpdateStatus
	Available           *Release
	History             []Release
	Downloaded, Total   int64
	Error               string
}
type UpdaterOptions struct {
	Version, FeedURL, PublicKey string
	Client                      *http.Client
	Layout                      *InstallLayout
	ExpectedRoot                func(version string) string
	OnChange                    func()
}
type Updater struct {
	mu     sync.Mutex
	opts   UpdaterOptions
	key    ed25519.PublicKey
	state  UpdateSnapshot
	staged *StagedUpdate
	busy   bool
	closed bool
	cancel context.CancelFunc
}
type StagedUpdate struct {
	Directory, Artifact string
	Release             Release
}

func NewUpdater(opts UpdaterOptions) *Updater {
	if opts.FeedURL == "" {
		opts.FeedURL = ReleaseFeedURL
	}
	if opts.PublicKey == "" {
		opts.PublicKey = ReleasePublicKey
	}
	key, _ := base64.StdEncoding.DecodeString(strings.TrimSpace(opts.PublicKey))
	if opts.Client == nil {
		opts.Client = &http.Client{Timeout: 10 * time.Minute}
	}
	// Reject redirects away from HTTPS, including the final GitHub asset URL.
	clone := *opts.Client
	previous := clone.CheckRedirect
	clone.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if err := validDownloadURL(req.URL.String()); err != nil {
			return err
		}
		if previous != nil {
			return previous(req, via)
		}
		if len(via) >= 10 {
			return errors.New("too many redirects")
		}
		return nil
	}
	opts.Client = &clone
	enabled := len(key) == ed25519.PublicKeySize && validDownloadURL(opts.FeedURL) == nil
	return &Updater{opts: opts, key: key, state: UpdateSnapshot{Enabled: enabled, CanInstall: enabled && opts.Layout != nil, Status: UpdateIdle}}
}
func (u *Updater) Snapshot() UpdateSnapshot {
	u.mu.Lock()
	defer u.mu.Unlock()
	s := u.state
	s.History = append([]Release(nil), s.History...)
	if s.Available != nil {
		copy := *s.Available
		s.Available = &copy
	}
	return s
}
func (u *Updater) changed() {
	if u.opts.OnChange != nil {
		u.opts.OnChange()
	}
}
func (u *Updater) fail(err error) error {
	u.mu.Lock()
	u.state.Status = UpdateFailed
	u.state.Error = err.Error()
	u.mu.Unlock()
	u.changed()
	return err
}
func (u *Updater) begin(status UpdateStatus) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.closed {
		return os.ErrClosed
	}
	if !u.state.Enabled {
		return ErrUpdatesDisabled
	}
	if u.busy {
		return errors.New("an update operation is already running")
	}
	u.busy = true
	u.state.Status = status
	u.state.Error = ""
	return nil
}
func (u *Updater) end() {
	u.mu.Lock()
	if u.cancel != nil {
		u.cancel()
		u.cancel = nil
	}
	u.busy = false
	u.mu.Unlock()
}
func (u *Updater) operationContext(parent context.Context) context.Context {
	ctx, cancel := context.WithCancel(parent)
	u.mu.Lock()
	if u.closed {
		cancel()
	} else {
		u.cancel = cancel
	}
	u.mu.Unlock()
	return ctx
}

// Check fetches the signed appcast and version history without downloading or
// installing anything. It belongs on a worker goroutine, not a UI frame.
func (u *Updater) Check(ctx context.Context) error {
	if err := u.begin(UpdateChecking); err != nil {
		return err
	}
	defer u.end()
	ctx = u.operationContext(ctx)
	u.changed()
	raw, err := u.fetch(ctx, u.opts.FeedURL, MaxFeedBytes)
	if err != nil {
		return u.fail(err)
	}
	items, err := ParseAppcast(raw)
	if err != nil {
		return u.fail(err)
	}
	if len(items) == 0 {
		return u.fail(errors.New("the update feed has no signed release"))
	}
	u.mu.Lock()
	if u.closed {
		u.mu.Unlock()
		return os.ErrClosed
	}
	u.state.History = items
	u.state.Available = nil
	u.state.Status = UpdateIdle
	if CompareVersions(items[0].Version, u.opts.Version) > 0 {
		release := items[0]
		u.state.Available = &release
		u.state.Status = UpdateAvailable
	}
	u.mu.Unlock()
	u.changed()
	return nil
}

// Stage verifies the exact artifact size and Ed25519 signature before opening
// the archive. Its sibling directory makes the eventual install an atomic swap.
func (u *Updater) Stage(ctx context.Context) error {
	if err := u.begin(UpdateDownloading); err != nil {
		return err
	}
	defer u.end()
	ctx = u.operationContext(ctx)
	u.mu.Lock()
	release := u.state.Available
	layout := u.opts.Layout
	u.mu.Unlock()
	if release == nil {
		return u.fail(errors.New("no update is available"))
	}
	if layout == nil {
		return u.fail(errors.New("this install is check-only; no managed installation was found"))
	}
	r := *release
	if err := validateRelease(r); err != nil {
		return u.fail(err)
	}
	u.mu.Lock()
	u.state.Downloaded = 0
	u.state.Total = r.Length
	u.mu.Unlock()
	u.changed()
	dir, err := os.MkdirTemp(filepath.Dir(layout.Root), "."+filepath.Base(layout.Root)+".update-")
	if err != nil {
		return u.fail(err)
	}
	keep := false
	defer func() {
		if !keep {
			_ = os.RemoveAll(dir)
		}
	}()
	artifact := filepath.Join(dir, "payload.download")
	if err = u.download(ctx, r.URL, artifact, r.Length); err != nil {
		return u.fail(err)
	}
	if err = VerifyArtifact(artifact, r, u.key); err != nil {
		return u.fail(err)
	}
	if runtime.GOOS != "windows" {
		extracted := filepath.Join(dir, "contents")
		if err = os.Mkdir(extracted, 0700); err != nil {
			return u.fail(err)
		}
		root := DefaultArchiveRoot(r.Version, layout)
		if u.opts.ExpectedRoot != nil {
			root = u.opts.ExpectedRoot(r.Version)
		}
		if err = ExtractReleaseArchive(artifact, extracted, root); err != nil {
			return u.fail(err)
		}
		if err = validateExecutable(extracted, layout.Executable); err != nil {
			return u.fail(err)
		}
		_ = os.Remove(artifact)
		// Move contents into a namespaced sibling rather than leaving the archive's
		// outer directory in the install tree.
		staging, err := os.MkdirTemp(filepath.Dir(layout.Root), "."+filepath.Base(layout.Root)+".update-")
		if err != nil {
			return u.fail(err)
		}
		_ = os.Remove(staging)
		if err = os.Rename(extracted, staging); err != nil {
			return u.fail(err)
		}
		_ = os.RemoveAll(dir)
		dir = staging
		artifact = ""
	} else {
		target := filepath.Join(dir, "Orbit-Setup.exe")
		if err = os.Rename(artifact, target); err != nil {
			return u.fail(err)
		}
		artifact = target
	}
	u.mu.Lock()
	if u.closed {
		u.mu.Unlock()
		return os.ErrClosed
	}
	old := u.staged
	u.staged = &StagedUpdate{Directory: dir, Artifact: artifact, Release: r}
	u.state.Status = UpdateStaged
	u.mu.Unlock()
	keep = true
	if old != nil {
		_ = os.RemoveAll(old.Directory)
	}
	u.changed()
	return nil
}
func (u *Updater) Close() error {
	u.mu.Lock()
	u.closed = true
	if u.cancel != nil {
		u.cancel()
		u.cancel = nil
	}
	s := u.staged
	u.staged = nil
	u.mu.Unlock()
	if s != nil {
		return os.RemoveAll(s.Directory)
	}
	return nil
}
func (u *Updater) fetch(ctx context.Context, raw string, limit int64) ([]byte, error) {
	if err := validDownloadURL(raw); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Orbit-MyGo/"+u.opts.Version)
	resp, err := u.opts.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update request returned HTTP %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err == nil && int64(len(b)) > limit {
		err = errors.New("update response exceeds size limit")
	}
	return b, err
}
func (u *Updater) download(ctx context.Context, raw, dest string, length int64) error {
	if err := validDownloadURL(raw); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", raw, nil)
	if err != nil {
		return err
	}
	resp, err := u.opts.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update download returned HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > 0 && resp.ContentLength != length {
		return errors.New("artifact Content-Length differs from signed feed")
	}
	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	reader := io.LimitReader(resp.Body, length+1)
	buf := make([]byte, 64<<10)
	var total int64
	for {
		n, e := reader.Read(buf)
		if n > 0 {
			total += int64(n)
			if total > length {
				return errors.New("artifact exceeds declared length")
			}
			if _, err = f.Write(buf[:n]); err != nil {
				return err
			}
			u.mu.Lock()
			u.state.Downloaded = total
			u.mu.Unlock()
			u.changed()
		}
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
	}
	if total != length {
		return errors.New("artifact differs from declared length")
	}
	return f.Sync()
}
func validDownloadURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" {
		return errors.New("updates require an HTTPS URL without credentials or fragment")
	}
	return nil
}
func validateRelease(r Release) error {
	if r.Version == "" || len(r.Version) > 64 {
		return errors.New("invalid update version")
	}
	for _, c := range r.Version {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune(".-+_", c)) {
			return errors.New("invalid update version")
		}
	}
	if r.Length <= 0 || r.Length > MaxArtifactBytes {
		return errors.New("invalid artifact length")
	}
	return validDownloadURL(r.URL)
}
func VerifyArtifact(filename string, r Release, key ed25519.PublicKey) error {
	if len(key) != ed25519.PublicKeySize {
		return errors.New("invalid update public key")
	}
	sig, err := base64.StdEncoding.DecodeString(strings.TrimSpace(r.Signature))
	if err != nil || len(sig) != ed25519.SignatureSize {
		return errors.New("malformed update signature")
	}
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	if info.Size() != r.Length || info.Size() > MaxArtifactBytes {
		return errors.New("artifact length mismatch")
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if !ed25519.Verify(key, data, sig) {
		return errors.New("update artifact failed signature verification")
	}
	return nil
}

func ParseAppcast(data []byte) ([]Release, error) {
	if int64(len(data)) > MaxFeedBytes {
		return nil, errors.New("appcast exceeds size limit")
	}
	d := xml.NewDecoder(strings.NewReader(string(data)))
	var result []Release
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "item" {
			continue
		}
		var item struct {
			Version     string `xml:"shortVersionString"`
			Description string `xml:"description"`
			Enclosure   struct {
				Attrs []xml.Attr `xml:",any,attr"`
			} `xml:"enclosure"`
		}
		if err = d.DecodeElement(&item, &start); err != nil {
			return nil, err
		}
		r := Release{Version: strings.TrimSpace(item.Version), Notes: strings.TrimSpace(item.Description)}
		for _, a := range item.Enclosure.Attrs {
			switch a.Name.Local {
			case "url":
				r.URL = a.Value
			case "edSignature":
				r.Signature = a.Value
			case "shortVersionString":
				if r.Version == "" {
					r.Version = a.Value
				}
			case "length":
				r.Length, _ = strconv.ParseInt(a.Value, 10, 64)
			}
		}
		if r.Version != "" && r.URL != "" && r.Signature != "" {
			result = append(result, r)
		}
	}
	slices.SortStableFunc(result, func(a, b Release) int { return CompareVersions(b.Version, a.Version) })
	return result, nil
}

// CompareVersions preserves Orbit's dotted numeric ordering: prerelease/build
// suffixes are ignored, and missing fields are zero.
func CompareVersions(a, b string) int {
	parts := func(v string) []string {
		head := strings.SplitN(strings.SplitN(v, "-", 2)[0], "+", 2)[0]
		return strings.Split(head, ".")
	}
	if a == "" {
		a = "0"
	}
	if b == "" {
		b = "0"
	}
	aa, bb := parts(a), parts(b)
	for i := 0; i < max(len(aa), len(bb)); i++ {
		var x, y uint64
		if i < len(aa) {
			x, _ = strconv.ParseUint(strings.TrimSpace(aa[i]), 10, 64)
		}
		if i < len(bb) {
			y, _ = strconv.ParseUint(strings.TrimSpace(bb[i]), 10, 64)
		}
		if x < y {
			return -1
		}
		if x > y {
			return 1
		}
	}
	return 0
}

// ExtractReleaseArchive deliberately rejects links, devices, duplicate paths,
// traversal, wrong-platform roots and oversized payloads, as Orbit's original
// updater does. A valid signature does not make filesystem paths safe.
func ExtractReleaseArchive(filename, destination, expectedRoot string) error {
	if expectedRoot == "" || strings.ContainsAny(expectedRoot, "/\\") {
		return errors.New("invalid expected archive root")
	}
	if info, err := os.Lstat(destination); err == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("archive destination is not a real directory")
		}
		entries, err := os.ReadDir(destination)
		if err != nil {
			return err
		}
		if len(entries) != 0 {
			return errors.New("archive destination is not empty")
		}
	} else if !os.IsNotExist(err) {
		return err
	} else if err = os.MkdirAll(destination, 0700); err != nil {
		return err
	}
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	seen := map[string]bool{}
	var total int64
	count := 0
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		count++
		if count > MaxArchiveEntries {
			return errors.New("too many archive entries")
		}
		name := strings.TrimSuffix(h.Name, "/")
		if name == "" || strings.ContainsAny(name, "\\\x00") || path.IsAbs(name) {
			return errors.New("unsafe archive path")
		}
		components := strings.Split(name, "/")
		if components[0] != expectedRoot {
			return errors.New("archive has an unexpected top-level directory")
		}
		for _, c := range components {
			if c == "" || c == "." || c == ".." {
				return errors.New("unsafe archive path")
			}
		}
		if h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeDir {
			return errors.New("archive contains a link or special file")
		}
		if h.Size < 0 || h.Size > MaxUnpackedBytes-total {
			return errors.New("archive exceeds unpacked size limit")
		}
		total += h.Size
		if len(components) == 1 {
			if h.Typeflag != tar.TypeDir {
				return errors.New("archive root is not a directory")
			}
			continue
		}
		relative := strings.Join(components[1:], "/")
		if seen[relative] {
			return errors.New("duplicate archive path")
		}
		seen[relative] = true
		target := filepath.Join(destination, filepath.FromSlash(relative))
		if h.Typeflag == tar.TypeDir {
			if err = os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(h.Mode)&0777)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, tr)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	if count == 0 {
		return errors.New("archive is empty")
	}
	return nil
}
