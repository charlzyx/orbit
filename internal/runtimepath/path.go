// Package runtimepath finds Pi/Node executables for GUI launches without
// importing the native UI. Explicit environment variants never inherit the
// parent HOME, PATH, credentials or overrides.
package runtimepath

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

var systemDirs = []string{"/opt/homebrew/bin", "/usr/local/bin", "/opt/local/bin", "/usr/bin", "/bin"}
var homeDirs = []string{".volta/bin", ".local/share/mise/shims", ".asdf/shims", ".local/bin", "Library/pnpm/bin", ".bun/bin", ".npm-global/bin", ".pi/agent/bin", ".yarn/bin"}
var versionDirs = [][2]string{{".nvm/versions/node", "bin"}, {".local/share/fnm/node-versions", "installation/bin"}}

func splitPath(value, goos string) []string {
	if goos == "windows" && runtime.GOOS != "windows" {
		return strings.Split(value, ";")
	}
	return filepath.SplitList(value)
}
func separator(goos string) string {
	if goos == "windows" {
		return ";"
	}
	return ":"
}

// Directories returns PATH, known system dirs, home-relative shims and real
// nvm/fnm version dirs (newest first). Empty home does not probe the cwd.
func Directories(path, home, goos string) []string {
	dirs := append([]string(nil), splitPath(path, goos)...)
	if goos != "windows" {
		dirs = append(dirs, systemDirs...)
	}
	if home == "" {
		return dirs
	}
	for _, sub := range homeDirs {
		dirs = append(dirs, filepath.Join(home, filepath.FromSlash(sub)))
	}
	for _, group := range versionDirs {
		base := filepath.Join(home, filepath.FromSlash(group[0]))
		children, _ := os.ReadDir(base)
		slices.SortFunc(children, func(a, b os.DirEntry) int {
			return compareVersion(strings.TrimPrefix(b.Name(), "v"), strings.TrimPrefix(a.Name(), "v"))
		})
		for _, child := range children {
			if child.IsDir() {
				dirs = append(dirs, filepath.Join(base, child.Name(), filepath.FromSlash(group[1])))
			}
		}
	}
	return dirs
}
func environmentValue(env []string, key, goos string) string {
	for i := len(env) - 1; i >= 0; i-- {
		name, value, ok := strings.Cut(env[i], "=")
		if ok && (name == key || goos == "windows" && strings.EqualFold(name, key)) {
			return value
		}
	}
	return ""
}
func directoriesIn(env []string, goos string) []string {
	home := environmentValue(env, "HOME", goos)
	if goos == "windows" {
		home = environmentValue(env, "USERPROFILE", goos)
	}
	dirs := Directories(environmentValue(env, "PATH", goos), home, goos)
	if goos == "windows" {
		for _, pair := range [][2]string{{"APPDATA", "npm"}, {"ProgramFiles", "nodejs"}, {"ProgramFiles", "Git/cmd"}, {"LOCALAPPDATA", "Programs/nodejs"}} {
			if base := environmentValue(env, pair[0], goos); base != "" {
				dirs = append(dirs, filepath.Join(base, filepath.FromSlash(pair[1])))
			}
		}
	}
	return dirs
}
func ExistingIn(dir, name, goos string) string {
	names := []string{name}
	if goos == "windows" && filepath.Ext(name) == "" {
		names = []string{name + ".exe", name + ".cmd", name + ".bat", name}
	}
	for _, name := range names {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			if goos == "windows" || info.Mode()&0111 != 0 {
				return path
			}
		}
	}
	return ""
}
func Resolve(name string) string { return ResolveIn(name, os.Environ(), runtime.GOOS) }
func ResolveIn(name string, env []string, goos string) string {
	if name == "pi" {
		if override := environmentValue(env, "PI_BIN", goos); override != "" {
			if info, err := os.Stat(override); err == nil && !info.IsDir() && (goos == "windows" || info.Mode()&0111 != 0) {
				return override
			}
			if !strings.ContainsAny(override, "/\\") {
				for _, dir := range splitPath(environmentValue(env, "PATH", goos), goos) {
					if dir != "" {
						if path := ExistingIn(dir, override, goos); path != "" {
							return path
						}
					}
				}
			}
			return ""
		}
	}
	for _, dir := range directoriesIn(env, goos) {
		if dir != "" {
			if path := ExistingIn(dir, name, goos); path != "" {
				return path
			}
		}
	}
	return ""
}
func augmentedPath(env []string, preferred, goos string) string {
	dirs := directoriesIn(env, goos)
	if preferred != "" {
		dirs = append([]string{preferred}, dirs...)
	}
	var out []string
	seen := map[string]bool{}
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		key := filepath.Clean(dir)
		if goos == "windows" {
			key = strings.ToLower(key)
		}
		if !seen[key] {
			seen[key] = true
			out = append(out, dir)
		}
	}
	return strings.Join(out, separator(goos))
}
func AugmentedPath(preferred string) string {
	return augmentedPath(os.Environ(), preferred, runtime.GOOS)
}

// AugmentEnvironment changes only PATH in the supplied environment. All other
// entries remain exactly as supplied, and no ambient variables are copied.
func AugmentEnvironment(base []string, preferred string) []string {
	return AugmentEnvironmentFor(base, preferred, runtime.GOOS)
}
func AugmentEnvironmentFor(base []string, preferred, goos string) []string {
	out := make([]string, 0, len(base)+1)
	for _, entry := range base {
		key, _, _ := strings.Cut(entry, "=")
		if key != "PATH" && !(goos == "windows" && strings.EqualFold(key, "PATH")) {
			out = append(out, entry)
		}
	}
	return append(out, "PATH="+augmentedPath(base, preferred, goos))
}
func compareVersion(a, b string) int {
	aa, bb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < max(len(aa), len(bb)); i++ {
		var x, y uint64
		if i < len(aa) {
			x, _ = strconv.ParseUint(aa[i], 10, 64)
		}
		if i < len(bb) {
			y, _ = strconv.ParseUint(bb[i], 10, 64)
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
