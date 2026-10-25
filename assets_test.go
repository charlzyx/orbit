package orbitassets

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestOriginalCatalogsAndResources(t *testing.T) {
	en, err := Locale("en")
	if err != nil {
		t.Fatal(err)
	}
	if len(en) < 2000 {
		t.Fatalf("source catalog changed: %d keys", len(en))
	}
	for _, lang := range Languages {
		catalog, err := Locale(lang)
		if err != nil {
			t.Fatal(err)
		}
		for key := range en {
			if _, ok := catalog[key]; !ok {
				t.Errorf("%s missing %s", lang, key)
			}
		}
	}
	for _, name := range []string{"app-icon.png", "fonts/zed/IBMPlexSans-Regular.ttf", "fonts/zed/Lilex-Regular.ttf", "icons/lock.svg"} {
		if b, err := Asset(name); err != nil || len(b) == 0 {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := fs.ReadFile(Extensions(), "orbit-guard-extension/index.js"); err != nil {
		t.Fatal(err)
	}
}

func TestTranslationFallbackAndPathGuard(t *testing.T) {
	if _, err := Asset("../../secret"); err == nil {
		t.Fatal("asset path escaped")
	}
	if _, err := Locale("../../secret"); err == nil {
		t.Fatal("locale path escaped")
	}
	got := Translate(nil, map[string]string{"hello": "Hello %{name}"}, "hello", map[string]string{"name": "Ada"})
	if got != "Hello Ada" {
		t.Fatal(got)
	}
}

// Literal calls to Tr never have a fallback. Check actual native callsites,
// rather than catalog-to-catalog completeness alone, to prevent raw IDs leaking.
func TestNativeLiteralTranslationKeysResolve(t *testing.T) {
	en, err := Locale("en")
	if err != nil {
		t.Fatal(err)
	}
	err = filepath.WalkDir("internal", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(f, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			fn, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || fn.Sel.Name != "Tr" {
				return true
			}
			literal, ok := call.Args[0].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			key, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Error(err)
				return true
			}
			if _, ok := en[key]; !ok {
				t.Errorf("%s: untranslated literal Tr key %q", path, key)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
