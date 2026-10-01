package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitAndAdd(t *testing.T) {
	d := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"init"}); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"add", "Example", "github://owner/repo"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(d, "components.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(d, "components", "example", "component.json")); err != nil {
		t.Fatal(err)
	}
}
