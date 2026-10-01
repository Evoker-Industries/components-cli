package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Evoker-Industries/components-cli/internal/util"
)

func Load(path string) (*Registry, error) {
	abs, err := util.MustAbs(path)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	var rf RegistryFile
	if err := json.Unmarshal(b, &rf); err != nil {
		return nil, fmt.Errorf("%s: invalid JSON: %w", abs, err)
	}
	components := map[string]ComponentRef{}
	for k, v := range rf.Components {
		components[k] = v
	}
	for k, v := range rf.LegacyComponents {
		if _, exists := components[k]; exists {
			return nil, fmt.Errorf("%s: duplicate component %q across components/Components", abs, k)
		}
		components[k] = v
	}
	forges := rf.Forges
	if forges == nil {
		forges = map[string]ForgeConfig{}
	}
	return &Registry{Path: abs, Dir: filepath.Dir(abs), Components: components, Forges: forges}, nil
}

func InitEmpty(path string, force bool) error {
	abs, err := util.MustAbs(path)
	if err != nil {
		return err
	}
	if _, err := os.Stat(abs); err == nil && !force {
		return fmt.Errorf("registry file already exists: %s", abs)
	}
	rf := RegistryFile{Components: map[string]ComponentRef{}, Forges: map[string]ForgeConfig{}}
	return util.WriteJSONAtomic(abs, rf, 0o644)
}

func Save(path string, reg *Registry) error {
	rf := RegistryFile{Components: reg.Components, Forges: reg.Forges}
	return util.WriteJSONAtomic(path, rf, 0o644)
}

func ResolveRef(reg *Registry, ref string) string {
	if filepath.IsAbs(ref) {
		return filepath.Clean(ref)
	}
	return filepath.Clean(filepath.Join(reg.Dir, ref))
}

func Validate(reg *Registry) error {
	seen := map[string]struct{}{}
	for name, c := range reg.Components {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return fmt.Errorf("%s: component name is empty", reg.Path)
		}
		if _, exists := seen[trimmed]; exists {
			return fmt.Errorf("%s: duplicate component name %q", reg.Path, trimmed)
		}
		seen[trimmed] = struct{}{}
		if strings.TrimSpace(c.Ref) == "" {
			return fmt.Errorf("%s: component %q missing ref", reg.Path, name)
		}
		manifestPath := ResolveRef(reg, c.Ref)
		if _, err := os.Stat(manifestPath); err != nil {
			return fmt.Errorf("%s: component %q manifest not found: %s", reg.Path, name, manifestPath)
		}
	}
	for alias, forge := range reg.Forges {
		if strings.TrimSpace(alias) == "" {
			return fmt.Errorf("%s: forge alias is empty", reg.Path)
		}
		if forge.Type == "" {
			return fmt.Errorf("%s: forge %q missing type", reg.Path, alias)
		}
		if forge.URL == "" {
			return fmt.Errorf("%s: forge %q missing url", reg.Path, alias)
		}
		t := strings.ToLower(strings.TrimSpace(forge.Type))
		switch t {
		case "github", "gitlab", "forgejo", "gitea":
		default:
			return fmt.Errorf("%s: forge %q has unsupported type %q", reg.Path, alias, forge.Type)
		}
		u, err := url.Parse(forge.URL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("%s: forge %q has invalid url %q", reg.Path, alias, forge.URL)
		}
		if forge.Auth != nil && strings.TrimSpace(forge.Auth.Secret) == "" {
			return fmt.Errorf("%s: forge %q has invalid auth.secret", reg.Path, alias)
		}
	}
	return nil
}
