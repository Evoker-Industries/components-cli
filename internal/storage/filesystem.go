package storage

import (
	"errors"
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

func (fs FS) BeginStaging(name, version string) (string, error) {
	stageRoot := filepath.Join(fs.Root, ".staging", SafeName(name))
	if err := os.MkdirAll(stageRoot, 0o755); err != nil {
		return "", err
	}
	return os.MkdirTemp(stageRoot, version+"-")
}

func (fs FS) PromoteStaging(stagingPath, name, version string) (string, error) {
	dest := fs.VersionPath(name, version)
	if _, err := os.Stat(dest); err == nil {
		return "", fmt.Errorf("version already installed: %s", dest)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	if err := os.Rename(stagingPath, dest); err != nil {
		return "", err
	}
	return dest, nil
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
