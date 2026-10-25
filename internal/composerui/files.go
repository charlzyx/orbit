package composerui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ScanResult struct {
	Files   []string
	Err     error
	Skipped int
	Capped  bool
}
type ScanFunc func(context.Context, string) ScanResult

var skippedDirectories = map[string]bool{".git": true, ".pi": true, ".hg": true, ".svn": true, "node_modules": true, "target": true, "dist": true, "build": true, "vendor": true, "__pycache__": true, ".next": true, ".cache": true}

// ScanWorkspace is the exact bounded source walk, with cancellation and explicit
// partial-error diagnostics added. It never reads file contents or follows dirs
// through symbolic links. A symlink itself is a source-compatible file entry.
func ScanWorkspace(ctx context.Context, root string) ScanResult {
	type dir struct {
		path  string
		depth int
	}
	result := ScanResult{Files: []string{}}
	if root == "" {
		return result
	}
	stack := []dir{{root, 0}}
	for len(stack) > 0 {
		if err := ctx.Err(); err != nil {
			result.Err = err
			return result
		}
		d := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if len(result.Files) >= MaxFiles {
			result.Capped = true
			continue
		}
		if d.depth > MaxDepth {
			continue
		}
		entries, err := os.ReadDir(d.path)
		if err != nil {
			result.Skipped++
			result.Err = errors.Join(result.Err, fmt.Errorf("read completion directory %s: %w", d.path, err))
			continue
		}
		for _, entry := range entries {
			if err := ctx.Err(); err != nil {
				result.Err = err
				return result
			}
			name := entry.Name()
			if strings.HasPrefix(name, ".") {
				continue
			}
			path := filepath.Join(d.path, name)
			if entry.IsDir() {
				if !skippedDirectories[name] {
					stack = append(stack, dir{path, d.depth + 1})
				}
				continue
			}
			if len(result.Files) >= MaxFiles {
				result.Capped = true
				continue
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				result.Skipped++
				result.Err = errors.Join(result.Err, err)
				continue
			}
			result.Files = append(result.Files, strings.ReplaceAll(filepath.ToSlash(rel), "\\", "/"))
		}
	}
	sort.Strings(result.Files)
	return result
}
