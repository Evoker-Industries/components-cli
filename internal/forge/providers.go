package forge

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Evoker-Industries/components-cli/internal/config"
	"github.com/Evoker-Industries/components-cli/internal/downloader"
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

func listGitLike(ctx context.Context, source ResolvedSource) ([]Version, error) {
	innerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	versions, err := (downloader.GitResolver{}).List(innerCtx, source.CloneURL)
	if err != nil {
		return nil, err
	}
	out := make([]Version, 0, len(versions))
	for _, v := range versions {
		out = append(out, Version{Value: v.Kind + ":" + v.Name})
	}
	return out, nil
}

func resolveGitLike(ctx context.Context, source ResolvedSource, requested string) (ResolvedVersion, error) {
	innerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	value, commit, err := (downloader.GitResolver{}).Resolve(innerCtx, source.CloneURL, requested)
	if err != nil {
		return ResolvedVersion{}, err
	}
	return ResolvedVersion{Value: value, Commit: commit}, nil
}

func (a githubAdapter) ListVersions(ctx context.Context, source ResolvedSource) ([]Version, error) {
	return listGitLike(ctx, source)
}
func (a githubAdapter) ResolveVersion(ctx context.Context, source ResolvedSource, requested string) (ResolvedVersion, error) {
	return resolveGitLike(ctx, source, requested)
}
func (a gitlabAdapter) ListVersions(ctx context.Context, source ResolvedSource) ([]Version, error) {
	return listGitLike(ctx, source)
}
func (a gitlabAdapter) ResolveVersion(ctx context.Context, source ResolvedSource, requested string) (ResolvedVersion, error) {
	return resolveGitLike(ctx, source, requested)
}
func (a forgejoAdapter) ListVersions(ctx context.Context, source ResolvedSource) ([]Version, error) {
	return listGitLike(ctx, source)
}
func (a forgejoAdapter) ResolveVersion(ctx context.Context, source ResolvedSource, requested string) (ResolvedVersion, error) {
	return resolveGitLike(ctx, source, requested)
}
func (a giteaAdapter) ListVersions(ctx context.Context, source ResolvedSource) ([]Version, error) {
	return listGitLike(ctx, source)
}
func (a giteaAdapter) ResolveVersion(ctx context.Context, source ResolvedSource, requested string) (ResolvedVersion, error) {
	return resolveGitLike(ctx, source, requested)
}
