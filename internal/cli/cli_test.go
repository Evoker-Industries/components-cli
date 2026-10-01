package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeRegistryWithGitHubForge(t *testing.T, path string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var reg map[string]any
	if err := json.Unmarshal(b, &reg); err != nil {
		t.Fatal(err)
	}
	reg["forges"] = map[string]any{
		"github": map[string]any{
			"type": "github",
			"url":  "https://github.com",
			"api":  "https://api.github.com",
		},
	}
	out, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	out = append(out, '\n')
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatal(err)
	}
}

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

func TestInitRefusesOverwriteWithoutForce(t *testing.T) {
	d := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"init"}); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"init"}); err == nil {
		t.Fatal("expected init overwrite refusal")
	}
	if err := Run([]string{"init", "--force"}); err != nil {
		t.Fatal(err)
	}
}

func TestAddDuplicateRefused(t *testing.T) {
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
	if err := Run([]string{"add", "Example", "github://owner/repo2"}); err == nil {
		t.Fatal("expected duplicate rejection")
	}
}

func TestInstallWritesLockChecksumsAndOffline(t *testing.T) {
	d := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"init"}); err != nil {
		t.Fatal(err)
	}
	writeRegistryWithGitHubForge(t, filepath.Join(d, "components.json"))
	if err := Run([]string{"add", "--version", "1.2.3", "Example", "github://owner/repo"}); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"install", "Example"}); err != nil {
		t.Fatal(err)
	}
	lockBytes, err := os.ReadFile(filepath.Join(d, "components.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	var lock struct {
		Components map[string]struct {
			ResolvedVersion string `json:"resolvedVersion"`
			Files           []struct {
				Path   string `json:"path"`
				SHA256 string `json:"sha256"`
			} `json:"files"`
		} `json:"components"`
	}
	if err := json.Unmarshal(lockBytes, &lock); err != nil {
		t.Fatal(err)
	}
	c := lock.Components["Example"]
	if c.ResolvedVersion != "1.2.3" {
		t.Fatalf("expected resolved version 1.2.3, got %q", c.ResolvedVersion)
	}
	if len(c.Files) == 0 || c.Files[0].SHA256 == "" {
		t.Fatalf("expected lock checksums, got %+v", c.Files)
	}
	if err := Run([]string{"--offline", "install", "Example"}); err != nil {
		t.Fatal(err)
	}
}

func TestCheckFailsUnknownForgeAlias(t *testing.T) {
	d := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"init"}); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"add", "Example", "company://org/repo"}); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"check"}); err == nil {
		t.Fatal("expected unknown forge alias failure")
	}
}
