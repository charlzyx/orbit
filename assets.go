// Copyright 2026 Orbit contributors. Licensed under Apache-2.0.
// Package orbitassets embeds the original Orbit resources for the native Go app.
// Keeping the original paths preserves their accompanying attribution notices.
package orbitassets

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
)

//go:embed all:crates/orbit-pi/assets crates/orbit-pi/locales/*.yml all:contrib
var resources embed.FS

// Asset reads an original resource relative to crates/orbit-pi/assets.
func Asset(name string) ([]byte, error) {
	if !fs.ValidPath(name) {
		return nil, fmt.Errorf("invalid asset path %q", name)
	}
	return resources.ReadFile("crates/orbit-pi/assets/" + name)
}

// Assets exposes the embedded icons, fonts and images without a working directory.
func Assets() fs.FS {
	f, _ := fs.Sub(resources, "crates/orbit-pi/assets")
	return f
}

// Extensions contains the original bundled Pi extensions and compatibility patches.
func Extensions() fs.FS {
	f, _ := fs.Sub(resources, "contrib")
	return f
}

var Languages = []string{"en", "zh-CN", "ja", "ko", "es", "fr", "de", "pt-BR", "ru", "it"}

// Locale reads the repository's generated flat YAML string catalog. This is not
// a general YAML parser: accepting only the format emitted by gen_locales.py
// makes unexpected source changes fail a test rather than alter a translation.
func Locale(language string) (map[string]string, error) {
	valid := false
	for _, code := range Languages {
		if language == code {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("unsupported language %q", language)
	}
	raw, err := resources.ReadFile("crates/orbit-pi/locales/" + language + ".yml")
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for line := 1; scanner.Scan(); line++ {
		s := strings.TrimSpace(scanner.Text())
		if s == "" || strings.HasPrefix(s, "#") || strings.HasPrefix(s, "_version:") {
			continue
		}
		key, value, ok := strings.Cut(s, ": ")
		if !ok {
			return nil, fmt.Errorf("locale %s line %d: expected flat string", language, line)
		}
		decoded, err := strconv.Unquote(value)
		if err != nil {
			return nil, fmt.Errorf("locale %s line %d: %w", language, line, err)
		}
		if _, exists := result[key]; exists {
			return nil, fmt.Errorf("locale %s duplicate key %q", language, key)
		}
		result[key] = decoded
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// Translate falls back to the canonical English string, then the key itself.
// Variables use Orbit's %{name} placeholder syntax without evaluating markup.
func Translate(catalog, english map[string]string, key string, variables map[string]string) string {
	s, ok := catalog[key]
	if !ok {
		s, ok = english[key]
	}
	if !ok {
		return key
	}
	for name, value := range variables {
		s = strings.ReplaceAll(s, "%{"+name+"}", value)
	}
	return s
}
