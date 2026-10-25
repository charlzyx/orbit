package agent

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/pirpc"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/runtimepath"
)

var patchLock sync.Mutex
var extensionFiles = map[string][]string{
	"orbit-quota-extension":    {"index.js", "adapters.js"},
	"orbit-guard-extension":    {"index.js", "policy.js"},
	"orbit-title-extension":    {"index.js", "title.js"},
	"orbit-workflow-extension": {"index.js", "policy.js"},
}

func prepareExtensions(options Options) ([]string, ExtensionStatus) {
	status := ExtensionStatus{}
	var paths []string
	resources := options.Resources
	if resources == nil {
		resources = orbitassets.Extensions()
	}
	if !options.DisableExtensions {
		for _, name := range []string{"orbit-quota-extension", "orbit-guard-extension", "orbit-title-extension", "orbit-workflow-extension"} {
			directory := filepath.Join(options.Home, ".orbit-pi", strings.TrimPrefix(name, "orbit-"))
			err := materialize(resources, name, extensionFiles[name], directory)
			if err != nil {
				status.Errors = append(status.Errors, name+": "+err.Error())
				continue
			}
			paths = append(paths, filepath.Join(directory, "index.js"))
			switch name {
			case "orbit-quota-extension":
				status.Quota = true
			case "orbit-guard-extension":
				status.Guard = true
			case "orbit-title-extension":
				status.Title = true
			case "orbit-workflow-extension":
				status.Workflow = true
			}
		}
	}
	if !options.DisablePatches && !patchEnvironmentDisabled(options) {
		status.Patches = applyPatches(options, resources)
	}
	return paths, status
}
func materialize(resources fs.FS, source string, files []string, target string) error {
	for _, name := range files {
		data, err := fs.ReadFile(resources, source+"/"+name)
		if err != nil {
			return err
		}
		if err = writeIfChanged(filepath.Join(target, name), data); err != nil {
			return err
		}
	}
	return nil
}
func writeIfChanged(path string, data []byte) error {
	if current, err := os.ReadFile(path); err == nil && bytes.Equal(current, data) {
		return nil
	}
	return atomicWrite(path, data)
}
func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".orbit-*")
	if err != nil {
		return err
	}
	name := file.Name()
	defer os.Remove(name)
	if err = file.Chmod(0600); err == nil {
		_, err = file.Write(data)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(name, path)
}
func findPackageRoot(binary, home string) string {
	path, err := exec.LookPath(binary)
	if err == nil {
		binary = path
	}
	real, err := filepath.EvalSymlinks(binary)
	if err == nil {
		binary = real
	}
	directory := filepath.Dir(binary)
	adjacent := filepath.Join(directory, "node_modules", "@earendil-works", "pi-coding-agent")
	if _, err := os.Stat(filepath.Join(adjacent, "package.json")); err == nil {
		return adjacent
	}
	for count := 0; count < 16; count++ {
		var pkg struct{ Name string }
		if data, err := os.ReadFile(filepath.Join(directory, "package.json")); err == nil && json.Unmarshal(data, &pkg) == nil && pkg.Name == "@earendil-works/pi-coding-agent" {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			break
		}
		directory = parent
	}
	install := filepath.Join(home, ".pi", "agent", "install")
	if version, err := os.ReadFile(filepath.Join(install, "current-version")); err == nil {
		root := filepath.Join(install, "releases", strings.TrimSpace(string(version)), "node_modules", "@earendil-works", "pi-coding-agent")
		if _, err := os.Stat(filepath.Join(root, "package.json")); err == nil {
			return root
		}
	}
	return ""
}
func applyPatches(options Options, resources fs.FS) PatchStatus {
	patchLock.Lock()
	defer patchLock.Unlock()
	status := PatchStatus{Enabled: true}
	base := filepath.Join(options.Home, ".orbit-pi", "rpc-patches")
	config := preferences.New(options.Home).RPCPatches()
	if !config.Enabled {
		status.Enabled = false
		return status
	}
	binary := options.PiBinary
	if binary == "" {
		binary = pirpc.ResolveBinary()
	}
	root := config.PackageRoot
	if configured := options.EnvOverrides["PI_PACKAGE_ROOT"]; configured != "" {
		root = configured
	}
	if root == "" {
		root = findPackageRoot(binary, options.Home)
	}
	status.PackageRoot = root
	if root == "" {
		status.Error = "Pi package root not found"
		return status
	}
	node := runtimepath.Resolve("node")
	if node == "" {
		status.Error = "Node executable not found"
		return status
	}
	var err error

	patches := []struct {
		source, target string
		files          []string
	}{
		{"pi-custom-ui-rpc", "custom-ui-rpc", []string{"apply.mjs", "custom-ui-handler.js"}},
		{"pi-quota-rpc", "quota-rpc", []string{"apply.mjs", "quota-handler.js"}},
		{"pi-auth-rpc", "auth-rpc", []string{"apply.mjs"}},
	}
	fingerprint := sha256.New()
	for _, patch := range patches {
		for _, name := range patch.files {
			data, err := fs.ReadFile(resources, patch.source+"/"+name)
			if err != nil {
				status.Error = err.Error()
				return status
			}
			fingerprint.Write(data)
		}
	}
	key := hex.EncodeToString(fingerprint.Sum(nil))
	signature := bundleSignature(root)
	var prior struct{ Fingerprint, Signature string }
	if data, err := os.ReadFile(filepath.Join(base, "go-applied.json")); err == nil && json.Unmarshal(data, &prior) == nil && prior.Fingerprint == key && prior.Signature == signature && signature != "" {
		for _, patch := range patches {
			status.Applied = append(status.Applied, patch.target)
		}
		return status
	}
	// An interprocess lock prevents simultaneous app instances rewriting a bundle.
	lock := filepath.Join(base, "go-patch.lock")
	if err = os.MkdirAll(base, 0700); err != nil {
		status.Error = err.Error()
		return status
	}
	handle, err := os.OpenFile(lock, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		status.Error = "Pi patch pass already active or stale lock needs review"
		return status
	}
	handle.Close()
	defer os.Remove(lock)
	for _, patch := range patches {
		directory := filepath.Join(base, patch.target)
		if err = materialize(resources, patch.source, patch.files, directory); err != nil {
			status.Error = err.Error()
			return status
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		command := exec.CommandContext(ctx, node, filepath.Join(directory, "apply.mjs"))
		command.Env = append(os.Environ(), "PI_BIN="+binary, "PI_PACKAGE_ROOT="+root)
		output, err := command.CombinedOutput()
		cancel()
		if err != nil {
			status.Error = patch.target + ": " + truncate(string(output), 1000)
			return status
		}
		status.Applied = append(status.Applied, patch.target)
	}
	saved, _ := json.Marshal(struct{ Fingerprint, Signature string }{key, bundleSignature(root)})
	if err = atomicWrite(filepath.Join(base, "go-applied.json"), saved); err != nil {
		status.Error = err.Error()
	}
	return status
}
func bundleSignature(root string) string {
	var values []string
	_ = filepath.WalkDir(filepath.Join(root, "dist", "bundle"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".js") {
			if info, err := entry.Info(); err == nil {
				values = append(values, fmt.Sprintf("%s:%d:%d", path, info.Size(), info.ModTime().UnixNano()))
			}
		}
		return nil
	})
	if len(values) == 0 {
		return ""
	}
	sort.Strings(values)
	hash := sha256.Sum256([]byte(strings.Join(values, "\n")))
	return hex.EncodeToString(hash[:])
}
