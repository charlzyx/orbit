package native

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

const AppTempDirectory = "orbit-pi"
const AIScratchGrace = time.Hour
const maximumTempSizeDepth = 4

var piScratchPrefixes = []string{"pi-bash-", "pi-powershell-", "pi-mcp-", "pi-codemode-", "pi-editor-", "pi-share-", "pi-output-", "pi-tui-crash.log", "pi-tui-debug.log"}

type TempEntry struct {
	Path, Name string
	IsDir      bool
	Bytes      uint64
	Modified   time.Time
	Active     bool
}
type TempFiles struct {
	Directory          string
	Entries, AIEntries []TempEntry
	root               string
}
type CleanFailure struct{ Path, Error string }
type CleanOutcome struct {
	Removed       int
	Bytes         uint64
	SkippedActive []string
	Failed        []CleanFailure
}

func TempRoot() string      { return os.TempDir() }
func TempDirectory() string { return filepath.Join(TempRoot(), AppTempDirectory) }
func EnsureTempDir() (string, error) {
	dir := TempDirectory()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("Orbit temporary directory is not a real directory: %s", dir)
	}
	return dir, nil
}
func ScanTempFiles() TempFiles { return ScanTempFilesIn(TempRoot(), time.Now()) }

// ScanTempFilesIn exists so privacy behavior can be tested in a disposable
// tree. It never creates or removes anything and never follows symlinks.
func ScanTempFilesIn(root string, now time.Time) TempFiles {
	root, _ = filepath.Abs(root)
	if real, err := filepath.EvalSymlinks(root); err == nil {
		root = real
	}
	dir := filepath.Join(root, AppTempDirectory)
	files := TempFiles{Directory: dir, root: root}
	if info, err := os.Lstat(dir); err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
		if children, err := os.ReadDir(dir); err == nil {
			for _, child := range children {
				if entry, err := readTempEntry(filepath.Join(dir, child.Name()), false, now); err == nil {
					files.Entries = append(files.Entries, entry)
				}
			}
		}
	}
	if children, err := os.ReadDir(root); err == nil {
		for _, child := range children {
			name := child.Name()
			if name == AppTempDirectory {
				continue
			}
			ai := isPiScratch(name)
			if !strings.HasPrefix(name, "orbit-") && !ai {
				continue
			}
			if entry, err := readTempEntry(filepath.Join(root, name), ai, now); err == nil {
				if ai {
					files.AIEntries = append(files.AIEntries, entry)
				} else {
					files.Entries = append(files.Entries, entry)
				}
			}
		}
	}
	order := func(a, b TempEntry) int {
		if c := b.Modified.Compare(a.Modified); c != 0 {
			return c
		}
		return strings.Compare(b.Path, a.Path)
	}
	slices.SortFunc(files.Entries, order)
	slices.SortFunc(files.AIEntries, order)
	return files
}
func (f TempFiles) TotalBytes() uint64 {
	var total uint64
	for _, e := range f.Entries {
		total += e.Bytes
	}
	return total
}
func (f TempFiles) AITotalBytes() uint64 {
	var total uint64
	for _, e := range f.AIEntries {
		total += e.Bytes
	}
	return total
}
func isPiScratch(name string) bool {
	for _, prefix := range piScratchPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
func embeddedPID(name string) (int, bool) {
	var value string
	switch {
	case strings.HasPrefix(name, "bug-report-pid-"):
		value = strings.SplitN(strings.TrimPrefix(name, "bug-report-pid-"), "-", 2)[0]
	case strings.HasPrefix(name, "update-"):
		value = strings.SplitN(strings.TrimPrefix(name, "update-"), "-", 2)[0]
	case strings.HasPrefix(name, "login-") && strings.HasSuffix(name, ".command"):
		value = strings.TrimSuffix(strings.TrimPrefix(name, "login-"), ".command")
	case strings.HasPrefix(name, "orbit-update-"):
		value = strings.SplitN(strings.TrimPrefix(name, "orbit-update-"), "-", 2)[0]
	case strings.HasPrefix(name, "orbit-") && strings.HasSuffix(name, ".command"):
		value = strings.TrimSuffix(strings.TrimPrefix(name, "orbit-"), ".command")
	default:
		return 0, false
	}
	pid, err := strconv.ParseUint(value, 10, 32)
	return int(pid), err == nil && pid > 0 && pid <= math.MaxInt32
}
func tempEntryActive(name string) bool {
	pid, ok := embeddedPID(name)
	return ok && (pid == os.Getpid() || processAlive(pid))
}
func readTempEntry(path string, ai bool, now time.Time) (TempEntry, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return TempEntry{}, err
	}
	size, newest := tempPathSize(path, 0)
	active := tempEntryActive(filepath.Base(path))
	if ai && now.Sub(newest) < AIScratchGrace {
		active = true
	}
	return TempEntry{Path: path, Name: filepath.Base(path), IsDir: info.IsDir(), Bytes: size, Modified: info.ModTime(), Active: active}, nil
}
func tempPathSize(path string, depth int) (uint64, time.Time) {
	info, err := os.Lstat(path)
	if err != nil {
		return 0, time.Time{}
	}
	modified := info.ModTime()
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return uint64(max(0, info.Size())), modified
	}
	if depth >= maximumTempSizeDepth {
		return 0, modified
	}
	children, err := os.ReadDir(path)
	if err != nil {
		return 0, modified
	}
	var total uint64
	for _, child := range children {
		size, t := tempPathSize(filepath.Join(path, child.Name()), depth+1)
		total += size
		if t.After(modified) {
			modified = t
		}
	}
	return total, modified
}

// CleanTempFiles removes only entries in this scan and category, re-checking
// activity at the click, not trusting a stale page snapshot. All names and
// parents are revalidated; there is no broad pi-* cleanup.
func CleanTempFiles(files TempFiles, ai bool) CleanOutcome {
	return cleanTempFilesAt(files, ai, time.Now())
}
func cleanTempFilesAt(files TempFiles, ai bool, now time.Time) CleanOutcome {
	entries := files.Entries
	if ai {
		entries = files.AIEntries
	}
	out := CleanOutcome{}
	for _, entry := range entries {
		path := filepath.Clean(entry.Path)
		parent, name := filepath.Dir(path), filepath.Base(path)
		scoped := files.root != "" && files.Directory == filepath.Join(files.root, AppTempDirectory)
		if ai {
			scoped = scoped && parent == files.root && isPiScratch(name)
		} else if parent == files.root {
			scoped = scoped && name != AppTempDirectory && strings.HasPrefix(name, "orbit-")
		} else {
			scoped = scoped && parent == files.Directory
			if info, err := os.Lstat(files.Directory); err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				scoped = false
			}
		}
		if !scoped {
			out.Failed = append(out.Failed, CleanFailure{path, "entry is outside the selected Orbit scratch scope"})
			continue
		}
		current, err := readTempEntry(path, ai, now)
		if os.IsNotExist(err) {
			out.Removed++
			continue
		}
		if err != nil {
			out.Failed = append(out.Failed, CleanFailure{path, err.Error()})
			continue
		}
		if entry.Active || current.Active {
			out.SkippedActive = append(out.SkippedActive, path)
			continue
		}
		info, err := os.Lstat(path)
		if err == nil {
			if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
				err = os.RemoveAll(path)
			} else {
				err = os.Remove(path)
			}
		}
		if err == nil {
			out.Removed++
			out.Bytes += current.Bytes
		} else if os.IsNotExist(err) {
			out.Removed++
		} else {
			out.Failed = append(out.Failed, CleanFailure{path, err.Error()})
		}
	}
	return out
}
