package component

type VerifyPolicy struct {
	SHA256 bool `json:"sha256,omitempty"`
}

type Validation struct {
	RequiredFiles []string `json:"requiredFiles,omitempty"`
}

type UpdatePolicy struct{}

type RawComponent struct {
	Name       string        `json:"name"`
	Source     string        `json:"source"`
	Version    string        `json:"version,omitempty"`
	Ver        string        `json:"ver,omitempty"`
	Directory  string        `json:"directory,omitempty"`
	Dir        string        `json:"dir,omitempty"`
	Update     *UpdatePolicy `json:"update,omitempty"`
	Verify     *VerifyPolicy `json:"verify,omitempty"`
	Validation *Validation   `json:"validation,omitempty"`
}

type ManifestFile struct {
	Component RawComponent `json:"component"`
}

type Component struct {
	Name       string
	Source     string
	Version    string
	Directory  string
	Update     *UpdatePolicy
	Verify     *VerifyPolicy
	Validation *Validation
}
