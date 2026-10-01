package lockfile

import "time"

type File struct {
	Schema      string                   `json:"schema"`
	GeneratedAt time.Time                `json:"generatedAt"`
	Components  map[string]ComponentLock `json:"components"`
}

type FileView struct {
	Schema      string                   `json:"schema"`
	GeneratedAt string                   `json:"generatedAt"`
	Components  map[string]ComponentLock `json:"components"`
}

type ComponentLock struct {
	Source           string     `json:"source"`
	RequestedVersion string     `json:"requestedVersion,omitempty"`
	ResolvedVersion  string     `json:"resolvedVersion,omitempty"`
	Commit           string     `json:"commit,omitempty"`
	Path             string     `json:"path,omitempty"`
	Verified         bool       `json:"verified,omitempty"`
	Files            []FileHash `json:"files,omitempty"`
}

type FileHash struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
