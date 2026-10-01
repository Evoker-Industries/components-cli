package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func WriteJSONAtomic(path string, v any, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return WriteBytesAtomic(path, b, perm)
}

func WriteBytesAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	return nil
}

func EnsureGitIgnoreEntry(repoDir, entry string) error {
	if entry == "" {
		return nil
	}
	path := filepath.Join(repoDir, ".gitignore")
	var existing []byte
	if b, err := os.ReadFile(path); err == nil {
		existing = b
	} else if !os.IsNotExist(err) {
		return err
	}
	content := string(existing)
	if content != "" {
		for _, line := range splitLines(content) {
			if line == entry {
				return nil
			}
		}
	}
	if content != "" && content[len(content)-1] != '\n' {
		content += "\n"
	}
	content += entry + "\n"
	return os.WriteFile(path, []byte(content), 0o644)
}

func splitLines(s string) []string {
	out := []string{}
	cur := ""
	for _, r := range s {
		if r == '\n' {
			out = append(out, cur)
			cur = ""
			continue
		}
		if r != '\r' {
			cur += string(r)
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

func RelOrAbs(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	if rel == "." {
		return "./"
	}
	if rel == "" {
		return target
	}
	if rel[0] != '.' && rel[0] != '/' {
		return "./" + rel
	}
	return rel
}

func MustAbs(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", path, err)
	}
	return abs, nil
}
