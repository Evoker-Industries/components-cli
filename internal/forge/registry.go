package forge

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Evoker-Industries/components-cli/internal/config"
	"github.com/Evoker-Industries/components-cli/internal/downloader"
	"github.com/Evoker-Industries/components-cli/internal/source"
)

type ResolvedSource struct {
	Type       string
	Alias      string
	BaseURL    string
	APIURL     string
	Repository string
	CloneURL   string
	SecretRef  string
}

type Registry struct {
	adapters map[string]ForgeAdapter
}

func NewRegistry() *Registry {
	r := &Registry{adapters: map[string]ForgeAdapter{}}
	r.Register(NewGitHubAdapter())
	r.Register(NewGitLabAdapter())
	r.Register(NewForgejoAdapter())
	r.Register(NewGiteaAdapter())
	return r
}

func (r *Registry) Register(adapter ForgeAdapter) {
	r.adapters[adapter.Type()] = adapter
}

func (r *Registry) Adapter(t string) (ForgeAdapter, bool) {
	a, ok := r.adapters[strings.ToLower(t)]
	return a, ok
}

func (r *Registry) Resolve(uri *source.SourceURI, forges map[string]config.ForgeConfig) (*ResolvedSource, error) {
	scheme := strings.ToLower(uri.Scheme)
	target := uri.Target
	switch scheme {
	case "git+https", "git+ssh":
		return &ResolvedSource{Type: "git", Alias: scheme, Repository: target, CloneURL: scheme + "://" + target}, nil
	case "url":
		u, err := url.Parse(target)
		if err != nil || u.Scheme == "" {
			return nil, fmt.Errorf("invalid url source target %q", target)
		}
		return &ResolvedSource{Type: "url", Alias: "url", Repository: target, CloneURL: target}, nil
	case "npm", "unpkg", "pypi":
		return nil, fmt.Errorf("source scheme %q is not yet supported", scheme)
	}
	if forgeCfg, ok := forges[scheme]; ok {
		adapter, ok := r.Adapter(forgeCfg.Type)
		if !ok {
			return nil, fmt.Errorf("unsupported forge type %q for alias %q", forgeCfg.Type, scheme)
		}
		repoURL := adapter.RepositoryURL(forgeCfg, target)
		cloneURL := adapter.CloneURL(forgeCfg, target)
		return &ResolvedSource{
			Type:       strings.ToLower(forgeCfg.Type),
			Alias:      scheme,
			BaseURL:    strings.TrimRight(forgeCfg.URL, "/"),
			APIURL:     strings.TrimRight(forgeCfg.API, "/"),
			Repository: repoURL,
			CloneURL:   cloneURL,
			SecretRef: func() string {
				if forgeCfg.Auth == nil {
					return ""
				}
				return strings.TrimSpace(forgeCfg.Auth.Secret)
			}(),
		}, nil
	}
	return nil, fmt.Errorf("unknown forge alias or unsupported source scheme %q", scheme)
}

func (r *Registry) ResolveVersion(ctx context.Context, source *ResolvedSource, requested string) (ResolvedVersion, error) {
	switch source.Type {
	case "git":
		innerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		value, commit, err := (downloader.GitResolver{}).Resolve(innerCtx, source.CloneURL, requested)
		if err != nil {
			return ResolvedVersion{}, err
		}
		return ResolvedVersion{Value: value, Commit: commit}, nil
	case "url":
		if strings.TrimSpace(requested) == "" {
			return ResolvedVersion{Value: "latest"}, nil
		}
		return ResolvedVersion{Value: requested}, nil
	default:
		adapter, ok := r.Adapter(source.Type)
		if !ok {
			return ResolvedVersion{}, fmt.Errorf("unsupported source type %q", source.Type)
		}
		return adapter.ResolveVersion(ctx, *source, requested)
	}
}
