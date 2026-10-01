package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLegacyComponentsAndResolveRef(t *testing.T) {
	d := t.TempDir()
	manifest := filepath.Join(d, "x", "component.json")
	if err := os.MkdirAll(filepath.Dir(manifest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, []byte(`{"component":{"name":"X","source":"github://o/r"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	regFile := filepath.Join(d, "components.json")
	if err := os.WriteFile(regFile, []byte(`{"Components":{"X":{"ref":"./x/component.json"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	reg, err := Load(regFile)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reg.Components["X"]; !ok {
		t.Fatal("missing loaded component")
	}
	if got := ResolveRef(reg, reg.Components["X"].Ref); got != manifest {
		t.Fatalf("resolve ref mismatch: %s", got)
	}
}
