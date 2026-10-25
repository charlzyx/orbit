package appbundle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// The version key is the shipped manifest hash. Only public runtime resources
// are materialized, never normal-home configuration, credentials or sessions.
func (l Layout) prepareRuntime(home string) (string, error) {
	manifestPath := filepath.Join(l.Root, "Contents", "Resources", "orbit-bundle.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", fmt.Errorf("The bundled runtime manifest is unavailable: %w", err)
	}
	var manifest struct {
		Schema    int    `json:"schema"`
		PiVersion string `json:"pi_version"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil || manifest.Schema != 1 || manifest.PiVersion != "1.4.0" {
		return "", errors.New("The bundled runtime manifest is invalid or unsupported")
	}
	sum := sha256.Sum256(data)
	key := hex.EncodeToString(sum[:])
	base := filepath.Join(home, "runtime")
	target := filepath.Join(base, key)
	if err := rejectSymlinks(home, target); err != nil {
		return "", err
	}
	if info, err := os.Lstat(target); err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("Private runtime path is not a directory: %s", target)
		}
		return validateRuntime(target, key)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	if err := os.MkdirAll(base, 0700); err != nil {
		return "", err
	}
	stage, err := os.MkdirTemp(base, ".prepare-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(stage)
	source := filepath.Join(l.Root, "Contents", "Resources", "pi")
	if err := copyRuntime(source, filepath.Join(stage, "pi")); err != nil {
		return "", fmt.Errorf("Cannot prepare the private Pi runtime: %w", err)
	}
	if err := os.WriteFile(filepath.Join(stage, "ready.json"), []byte("{\"manifest_sha256\":\""+key+"\"}\n"), 0600); err != nil {
		return "", err
	}
	if _, err := validateRuntime(stage, key); err != nil {
		return "", err
	}
	if err := os.Rename(stage, target); err != nil {
		// A simultaneous first launch may have finished the same version. Never
		// replace a prior runtime; accept it only after checking its completion mark.
		if _, check := validateRuntime(target, key); check != nil {
			return "", err
		}
	}
	return validateRuntime(target, key)
}

func validateRuntime(root, key string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, "ready.json"))
	var ready struct {
		Manifest string `json:"manifest_sha256"`
	}
	if err != nil || json.Unmarshal(data, &ready) != nil || ready.Manifest != key {
		return "", fmt.Errorf("Private Pi runtime is incomplete: %s. In Finder, use Go → Go to Folder to locate this folder, rename it to keep a backup, then reopen Orbit. Your settings and sessions are kept separately", root)
	}
	pi := filepath.Join(root, "pi", "node_modules", ".bin", "pi")
	real, err := filepath.EvalSymlinks(pi)
	if err != nil || !inside(filepath.Join(root, "pi"), real) {
		return "", fmt.Errorf("Private Pi launcher is missing or points outside its runtime: %s", pi)
	}
	info, err := os.Stat(real)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&0111 == 0 {
		return "", fmt.Errorf("Private Pi launcher is not executable: %s", pi)
	}
	return pi, nil
}

func copyRuntime(source, target string) error {
	// WalkDir never follows links. Rewrite links relative to their new location,
	// and reject every link whose resolved target escapes the public Pi tree.
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(target, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			real, err := filepath.EvalSymlinks(path)
			if err != nil || !inside(source, real) {
				return fmt.Errorf("Bundled Pi link escapes its runtime: %s", path)
			}
			targetRel, err := filepath.Rel(source, real)
			if err != nil {
				return err
			}
			link, err := filepath.Rel(filepath.Dir(dst), filepath.Join(target, targetRel))
			if err != nil {
				return err
			}
			return os.Symlink(link, dst)
		}
		if info.IsDir() {
			return os.Mkdir(dst, 0700)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("Unsupported bundled runtime file: %s", path)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		// The private copy is deliberately writable for Orbit's normal Pi RPC
		// patches and package-management flow. Executable bits are preserved.
		out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600|info.Mode().Perm()&0111)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}
