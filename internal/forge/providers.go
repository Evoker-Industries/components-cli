package forge

import (
	"context"
	"fmt"
	"strings"

	"github.com/Evoker-Industries/components-cli/internal/config"
)

type baseAdapter struct{ typ string }

func (b baseAdapter) Type() string { return b.typ }
func (b baseAdapter) ListVersions(context.Context, ResolvedSource) ([]Version, error) {
	return nil, fmt.Errorf("version listing not implemented for %s", b.typ)
}
func (b baseAdapter) ResolveVersion(context.Context, ResolvedSource, string) (ResolvedVersion, error) {
	return ResolvedVersion{}, fmt.Errorf("version resolution not implemented for %s", b.typ)
}

type githubAdapter struct{ baseAdapter }

type gitlabAdapter struct{ baseAdapter }

type forgejoAdapter struct{ baseAdapter }

type giteaAdapter struct{ baseAdapter }

func NewGitHubAdapter() ForgeAdapter  { return githubAdapter{baseAdapter{"github"}} }
func NewGitLabAdapter() ForgeAdapter  { return gitlabAdapter{baseAdapter{"gitlab"}} }
func NewForgejoAdapter() ForgeAdapter { return forgejoAdapter{baseAdapter{"forgejo"}} }
func NewGiteaAdapter() ForgeAdapter   { return giteaAdapter{baseAdapter{"gitea"}} }

func joinRepo(base, repo string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(repo, "/")
}

func (a githubAdapter) RepositoryURL(cfg config.ForgeConfig, repo string) string {
	return joinRepo(cfg.URL, repo)
}
func (a githubAdapter) CloneURL(cfg config.ForgeConfig, repo string) string {
	return joinRepo(cfg.URL, repo) + ".git"
}
func (a gitlabAdapter) RepositoryURL(cfg config.ForgeConfig, repo string) string {
	return joinRepo(cfg.URL, repo)
}
func (a gitlabAdapter) CloneURL(cfg config.ForgeConfig, repo string) string {
	return joinRepo(cfg.URL, repo) + ".git"
}
func (a forgejoAdapter) RepositoryURL(cfg config.ForgeConfig, repo string) string {
	return joinRepo(cfg.URL, repo)
}
func (a forgejoAdapter) CloneURL(cfg config.ForgeConfig, repo string) string {
	return joinRepo(cfg.URL, repo) + ".git"
}
func (a giteaAdapter) RepositoryURL(cfg config.ForgeConfig, repo string) string {
	return joinRepo(cfg.URL, repo)
}
func (a giteaAdapter) CloneURL(cfg config.ForgeConfig, repo string) string {
	return joinRepo(cfg.URL, repo) + ".git"
}
