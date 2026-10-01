package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecretsSetListRemove(t *testing.T) {
	p := filepath.Join(t.TempDir(), ".secrets.json")
	if err := Init(p, true); err != nil {
		t.Fatal(err)
	}
	s, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set("github.token", "x"); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	if got := s.Keys(); len(got) != 1 || got[0] != "github.token" {
		t.Fatalf("unexpected keys: %v", got)
	}
	if err := s.Remove("github.token"); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm()&0o077 != 0 {
		t.Fatalf("permissions are not restrictive: %o", fi.Mode().Perm())
	}
}
