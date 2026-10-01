package forge

import (
	"context"

	"github.com/Evoker-Industries/components-cli/internal/config"
)

type Version struct{ Value string }

type ResolvedVersion struct {
	Value  string
	Commit string
}

type ForgeAdapter interface {
	Type() string
	RepositoryURL(config config.ForgeConfig, repository string) string
	CloneURL(config config.ForgeConfig, repository string) string
	ListVersions(ctx context.Context, source ResolvedSource) ([]Version, error)
	ResolveVersion(ctx context.Context, source ResolvedSource, requested string) (ResolvedVersion, error)
}
