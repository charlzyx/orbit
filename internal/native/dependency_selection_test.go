package native

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDependencyProbeUsesSelectedPiWithoutUnrelatedPathPi(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX fixture executable")
	}
	home := t.TempDir()
	bin := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin)
	t.Setenv("PI_BIN", "")
	marker := filepath.Join(home, "unrelated-was-run")
	scripts := map[string]string{"pi": "#!/bin/sh\nprintf ran > '" + marker + "'\nprintf unrelated", "node": "#!/bin/sh\nprintf v24.0.0", "git": "#!/bin/sh\nprintf git-fixture"}
	for name, body := range scripts {
		if err := os.WriteFile(filepath.Join(bin, name), []byte(body), 0755); err != nil {
			t.Fatal(err)
		}
	}
	selected := filepath.Join(t.TempDir(), "selected-pi")
	if err := os.WriteFile(selected, []byte("#!/bin/sh\nprintf selected-1.4.0"), 0755); err != nil {
		t.Fatal(err)
	}
	deps := CheckDependenciesFor(context.Background(), selected)
	if len(deps) != 3 || !deps[0].Installed || deps[0].Path != selected || deps[0].Version != "selected-1.4.0" || deps[0].InstallHint != "" {
		t.Fatalf("selected launcher mismatch: %+v", deps)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("queried unrelated PATH Pi", err)
	}
	deps = CheckDependenciesFor(context.Background(), filepath.Join(home, "missing-selected-pi"))
	if deps[0].Installed || deps[0].Error == "" || deps[0].InstallHint != "" {
		t.Fatal("missing custom launcher silently fell back/offered unrelated install", deps[0])
	}
	// The existing unqualified API keeps its PATH semantics; only the new
	// explicit selection API changes which executable is queried.
	deps = CheckDependencies(context.Background())
	if deps[0].Path != filepath.Join(bin, "pi") {
		t.Fatal("legacy dependency probe changed")
	}
}
