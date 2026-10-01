package forge

import (
	"testing"

	"github.com/Evoker-Industries/components-cli/internal/config"
	"github.com/Evoker-Industries/components-cli/internal/source"
)

func TestResolveForgeAlias(t *testing.T) {
	r := NewRegistry()
	u, _ := source.Parse("github://owner/repo")
	resolved, err := r.Resolve(u, map[string]config.ForgeConfig{"github": {Type: "github", URL: "https://github.com", API: "https://api.github.com"}})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.CloneURL != "https://github.com/owner/repo.git" {
		t.Fatalf("unexpected clone URL: %s", resolved.CloneURL)
	}
}

func TestResolveUnsupportedPackage(t *testing.T) {
	r := NewRegistry()
	u, _ := source.Parse("npm://x")
	if _, err := r.Resolve(u, nil); err == nil {
		t.Fatal("expected unsupported error")
	}
}
