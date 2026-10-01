package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var safeNameRe = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

type FS struct {
	Root string
}

func SafeName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = safeNameRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-.")
	if s == "" {
		s = "component"
	}
	return s
}

func (fs FS) ComponentPath(name string) string {
	return filepath.Join(fs.Root, SafeName(name))
}

func (fs FS) VersionPath(name, version string) string {
	return filepath.Join(fs.ComponentPath(name), version)
}

func (fs FS) InstallVersion(name, version string) (string, error) {
	vp := fs.VersionPath(name, version)
	if _, err := os.Stat(vp); err == nil {
		return "", fmt.Errorf("version already installed: %s", vp)
	}
	if err := os.MkdirAll(vp, 0o755); err != nil {
		return "", err
	}
	return vp, nil
}

func (fs FS) SetCurrent(name, version string) error {
	cp := fs.ComponentPath(name)
	if err := os.MkdirAll(cp, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(cp, "current"), []byte(version+"\n"), 0o644)
}

func (fs FS) Current(name string) (string, error) {
	b, err := os.ReadFile(filepath.Join(fs.ComponentPath(name), "current"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
