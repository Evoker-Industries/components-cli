package validation

import (
	"path/filepath"
	"runtime"
	"strings"
)

func IsSafeExtractPath(baseDir, archivePath string) bool {
	if strings.TrimSpace(archivePath) == "" {
		return false
	}
	if strings.ContainsRune(archivePath, '\x00') {
		return false
	}
	archivePath = strings.ReplaceAll(archivePath, "\\", "/")
	if filepath.IsAbs(archivePath) {
		return false
	}
	if runtime.GOOS != "windows" {
		if len(archivePath) >= 2 && archivePath[1] == ':' {
			return false
		}
	}
	cleaned := filepath.Clean(archivePath)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return false
	}
	joined := filepath.Join(baseDir, cleaned)
	rel, err := filepath.Rel(baseDir, joined)
	if err != nil {
		return false
	}
	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}
