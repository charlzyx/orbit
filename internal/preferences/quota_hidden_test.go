package preferences

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestHiddenQuotaProviderSourceFormatAndMalformedFallback(t *testing.T) {
	s := New(t.TempDir())
	if len(s.HiddenQuotaProviders()) != 0 {
		t.Fatal("missing file hid providers")
	}
	if err := os.MkdirAll(s.RootDir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(s.RootDir, HiddenQuotaProvidersFile)
	for _, raw := range []string{`{`, `{"providers":false}`, `{"providers":[]}`} {
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		if len(s.HiddenQuotaProviders()) != 0 {
			t.Fatal("malformed/empty document did not default", raw)
		}
	}
	if err := os.WriteFile(path, []byte(`{"providers":["beta",null,17,"","alpha","beta"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if got := s.HiddenQuotaProviders(); !reflect.DeepEqual(got, []string{"alpha", "beta"}) {
		t.Fatal("mixed source entries not normalized", got)
	}
	if err := s.SaveHiddenQuotaProviders([]string{"beta", "", "beta", "absent-from-current-catalog"}); err != nil {
		t.Fatal(err)
	}
	reload := New(s.Home)
	if got := reload.HiddenQuotaProviders(); !reflect.DeepEqual(got, []string{"absent-from-current-catalog", "beta"}) {
		t.Fatal("stable provider IDs did not survive reload", got)
	}
}
func TestHiddenQuotaProviderWriteFailureIsReported(t *testing.T) {
	s := New(t.TempDir())
	if err := os.MkdirAll(filepath.Join(s.RootDir, HiddenQuotaProvidersFile), 0700); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveHiddenQuotaProviders([]string{"alpha"}); err == nil {
		t.Fatal("directory collision reported persisted preference")
	}
}
