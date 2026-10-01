package component

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func Load(path string) (*Component, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var mf ManifestFile
	if err := json.Unmarshal(b, &mf); err != nil {
		return nil, fmt.Errorf("%s: invalid JSON: %w", path, err)
	}
	return Normalize(mf.Component, path)
}

func Normalize(raw RawComponent, filePath string) (*Component, error) {
	version := strings.TrimSpace(raw.Version)
	ver := strings.TrimSpace(raw.Ver)
	if version != "" && ver != "" && version != ver {
		return nil, fmt.Errorf("component %q: both \"ver\" and \"version\" are present with different values", raw.Name)
	}
	if version == "" {
		version = ver
	}
	directory := strings.TrimSpace(raw.Directory)
	dir := strings.TrimSpace(raw.Dir)
	if directory != "" && dir != "" && directory != dir {
		return nil, fmt.Errorf("component %q: both \"dir\" and \"directory\" are present with different values", raw.Name)
	}
	if directory == "" {
		directory = dir
	}
	if strings.TrimSpace(raw.Name) == "" {
		return nil, fmt.Errorf("%s: missing component.name", filePath)
	}
	if strings.TrimSpace(raw.Source) == "" {
		return nil, fmt.Errorf("%s: missing component.source", filePath)
	}
	if directory == "" {
		directory = "."
	}
	return &Component{
		Name:       raw.Name,
		Source:     raw.Source,
		Version:    version,
		Directory:  directory,
		Update:     raw.Update,
		Verify:     raw.Verify,
		Validation: raw.Validation,
	}, nil
}

func Save(path string, c *Component) error {
	mf := ManifestFile{Component: RawComponent{
		Name:       c.Name,
		Source:     c.Source,
		Version:    c.Version,
		Directory:  c.Directory,
		Update:     c.Update,
		Verify:     c.Verify,
		Validation: c.Validation,
	}}
	b, err := json.MarshalIndent(mf, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}
