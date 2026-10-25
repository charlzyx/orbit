// Package workspace implements Orbit's native filesystem and Git workbench.
// UI callers run I/O off their render thread and own confirmation presentation.
package workspace

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"
)

const MaxFileBytes int64 = 2 * 1024 * 1024
const MaxFileLines = 200000

var ErrReadOnly = errors.New("document is read-only")
var ErrConfirmationRequired = errors.New("confirmation required")
var ErrDestinationExists = errors.New("destination already exists")
var ErrInvalidName = errors.New("invalid file name")

// ValidateName validates a single component, never a relative path.
func ValidateName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" || name == "." || name == ".." || len(name) > 255 || strings.ContainsAny(name, "/\\\x00") || strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return "", ErrInvalidName
	}
	if runtime.GOOS == "windows" {
		if strings.ContainsAny(name, `<>:"|?*`) || strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") {
			return "", ErrInvalidName
		}
		stem := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
		if stem == "CON" || stem == "PRN" || stem == "AUX" || stem == "NUL" || (len(stem) == 4 && (strings.HasPrefix(stem, "COM") || strings.HasPrefix(stem, "LPT")) && stem[3] >= '1' && stem[3] <= '9') {
			return "", ErrInvalidName
		}
	}
	return name, nil
}
func CreateFile(parent, name string) (string, error) {
	n, err := ValidateName(name)
	if err != nil {
		return "", err
	}
	p := filepath.Join(parent, n)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		return "", err
	}
	return p, f.Close()
}
func CreateDirectory(parent, name string) (string, error) {
	n, err := ValidateName(name)
	if err != nil {
		return "", err
	}
	p := filepath.Join(parent, n)
	return p, os.Mkdir(p, 0777)
}
func RenamePath(target, name string) (string, error) {
	n, err := ValidateName(name)
	if err != nil {
		return "", err
	}
	dst := filepath.Join(filepath.Dir(target), n)
	if filepath.Clean(dst) == filepath.Clean(target) {
		if _, err = os.Lstat(target); err != nil {
			return "", err
		}
		return dst, nil
	}
	if _, err = os.Lstat(dst); err == nil {
		return "", ErrDestinationExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	// Renaming is intentionally not replacement. A destination appearing between
	// the check and syscall is a platform filesystem race, not an overwrite API.
	if err = os.Rename(target, dst); err != nil {
		return "", err
	}
	return dst, nil
}
func DeleteIsRecoverable() bool { return runtime.GOOS == "darwin" }
func DeletePath(ctx context.Context, path string, confirmed bool) error {
	if !confirmed {
		return ErrConfirmationRequired
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := os.Lstat(path); err != nil {
		return err
	}
	if runtime.GOOS == "darwin" {
		script := `on run argv
 tell application "Finder" to delete POSIX file (item 1 of argv)
end run`
		out, err := exec.CommandContext(ctx, "osascript", "-e", script, "--", path).CombinedOutput()
		if err != nil {
			return fmt.Errorf("move to Trash: %w: %s", err, strings.TrimSpace(string(out)))
		}
		return nil
	}
	return os.RemoveAll(path)
}

type FileMode string

const (
	FileCode     FileMode = "code"
	FileText     FileMode = "text"
	FileMarkdown FileMode = "markdown"
	FileImage    FileMode = "image"
	FileBinary   FileMode = "binary"
	FileTooLarge FileMode = "too_large"
)

type FileContent struct {
	Path       string   `json:"path"`
	Display    string   `json:"display"`
	Mode       FileMode `json:"mode"`
	Language   string   `json:"language,omitempty"`
	Text       string   `json:"text,omitempty"`
	Lines      []string `json:"lines,omitempty"`
	ImageBytes []byte   `json:"imageBytes,omitempty"`
	Bytes      int64    `json:"bytes"`
	Truncated  bool     `json:"truncated"`
	Lossy      bool     `json:"lossy"`
	Error      string   `json:"error,omitempty"`
}

func (c FileContent) Editable() bool {
	return c.Error == "" && !c.Truncated && !c.Lossy && (c.Mode == FileCode || c.Mode == FileText || c.Mode == FileMarkdown)
}
func IsMarkdown(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown", ".mdown", ".mkd":
		return true
	}
	return false
}
func IsImage(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".svg", ".tiff":
		return true
	}
	return false
}
func LanguageForPath(path string) string {
	name := strings.ToLower(filepath.Base(path))
	if name == "dockerfile" {
		return "dockerfile"
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".rs":
		return "rust"
	case ".go":
		return "go"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".json", ".jsonc":
		return "json"
	case ".yml", ".yaml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".html", ".htm":
		return "html"
	case ".css", ".scss":
		return "css"
	case ".sh", ".bash", ".zsh":
		return "shell"
	case ".c", ".h", ".cc", ".cpp", ".hpp":
		return "cpp"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".sql":
		return "sql"
	case ".swift":
		return "swift"
	}
	return ""
}
func LoadFile(path string) FileContent {
	c := FileContent{Path: path, Display: filepath.Base(path), Mode: FileText}
	fi, err := os.Stat(path)
	if err != nil {
		c.Error = err.Error()
		return c
	}
	if !fi.Mode().IsRegular() {
		c.Error = "not a regular file"
		return c
	}
	c.Bytes = fi.Size()
	if c.Bytes > MaxFileBytes {
		c.Mode = FileTooLarge
		return c
	}
	f, err := os.Open(path)
	if err != nil {
		c.Error = err.Error()
		return c
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, MaxFileBytes+1))
	if err != nil {
		c.Error = err.Error()
		return c
	}
	c.Bytes = int64(len(raw))
	if c.Bytes > MaxFileBytes {
		c.Mode = FileTooLarge
		return c
	}
	if IsImage(path) {
		c.Mode = FileImage
		c.ImageBytes = raw
		return c
	}
	sniff := raw
	if len(sniff) > 8000 {
		sniff = sniff[:8000]
	}
	if bytes.IndexByte(sniff, 0) >= 0 {
		c.Mode = FileBinary
		return c
	}
	c.Lossy = !utf8.Valid(raw)
	c.Text = strings.ToValidUTF8(string(raw), "�")
	c.Lines = strings.Split(c.Text, "\n")
	if strings.HasSuffix(c.Text, "\n") {
		c.Lines = c.Lines[:len(c.Lines)-1]
	}
	if len(c.Lines) > MaxFileLines {
		c.Lines = c.Lines[:MaxFileLines]
		c.Truncated = true
	}
	if IsMarkdown(path) {
		c.Mode = FileMarkdown
	} else if lang := LanguageForPath(path); lang != "" {
		c.Mode = FileCode
		c.Language = lang
	}
	return c
}

// SaveExisting never creates a deleted/renamed file, and rejects non-regular
// targets rather than blocking on a FIFO or modifying a device.
func SaveExisting(path, text string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.Mode().IsRegular() {
		return ErrReadOnly
	}
	// Atomic replacement keeps the previous bytes intact until the complete new
	// buffer is durable. A unique sibling avoids concurrent-tab temp collisions.
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".orbit-save-*")
	if err != nil {
		return saveDirectExisting(path, text)
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(fi.Mode().Perm()); err == nil {
		_, err = io.WriteString(tmp, text)
	}
	if err == nil {
		err = tmp.Sync()
	}
	closeErr := tmp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	// Never resurrect a file that disappeared before the save was committed.
	current, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !current.Mode().IsRegular() {
		return ErrReadOnly
	}
	if err = os.Rename(name, path); err != nil {
		return saveDirectExisting(path, text)
	}
	return nil
}
func saveDirectExisting(path, text string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	_, werr := io.WriteString(f, text)
	cerr := f.Close()
	if werr != nil {
		return werr
	}
	return cerr
}
