package downloader

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

type GitVersion struct {
	Name   string
	Commit string
	Kind   string
}

type GitResolver struct{}

func (GitResolver) Resolve(ctx context.Context, cloneURL, requested string) (resolvedVersion string, commit string, err error) {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		requested = "latest"
	}
	if strings.HasPrefix(requested, "commit:") {
		sha := strings.TrimSpace(strings.TrimPrefix(requested, "commit:"))
		if sha == "" {
			return "", "", fmt.Errorf("empty commit in requested version")
		}
		return requested, sha, nil
	}
	if strings.HasPrefix(requested, "branch:") {
		branch := strings.TrimSpace(strings.TrimPrefix(requested, "branch:"))
		sha, err := lsRemoteSingle(ctx, cloneURL, "refs/heads/"+branch)
		if err != nil {
			return "", "", err
		}
		return "branch:" + branch, sha, nil
	}
	if strings.HasPrefix(requested, "tag:") {
		tag := strings.TrimSpace(strings.TrimPrefix(requested, "tag:"))
		sha, err := findTag(ctx, cloneURL, tag)
		if err != nil {
			return "", "", err
		}
		return "tag:" + tag, sha, nil
	}
	switch requested {
	case "latest", "stable", "nightly":
		branch, err := defaultBranch(ctx, cloneURL)
		if err != nil {
			branch = "main"
		}
		sha, err := lsRemoteSingle(ctx, cloneURL, "refs/heads/"+branch)
		if err != nil {
			return "", "", err
		}
		return requested, sha, nil
	}
	if sha, err := findTag(ctx, cloneURL, requested); err == nil {
		return requested, sha, nil
	}
	if strings.HasPrefix(requested, "v") {
		plain := strings.TrimPrefix(requested, "v")
		if sha, err := findTag(ctx, cloneURL, plain); err == nil {
			return requested, sha, nil
		}
	} else {
		if sha, err := findTag(ctx, cloneURL, "v"+requested); err == nil {
			return requested, sha, nil
		}
	}
	if sha, err := lsRemoteSingle(ctx, cloneURL, "refs/heads/"+requested); err == nil {
		return "branch:" + requested, sha, nil
	}
	return "", "", fmt.Errorf("could not resolve requested version %q", requested)
}

func (GitResolver) List(ctx context.Context, cloneURL string) ([]GitVersion, error) {
	entries, err := lsRemote(ctx, cloneURL, "refs/tags/*", "refs/heads/*")
	if err != nil {
		return nil, err
	}
	versions := make([]GitVersion, 0, len(entries))
	for _, e := range entries {
		kind := "ref"
		name := e.Ref
		if strings.HasPrefix(name, "refs/tags/") {
			kind = "tag"
			name = strings.TrimPrefix(name, "refs/tags/")
		}
		if strings.HasPrefix(name, "refs/heads/") {
			kind = "branch"
			name = strings.TrimPrefix(name, "refs/heads/")
		}
		versions = append(versions, GitVersion{Name: name, Commit: e.SHA, Kind: kind})
	}
	sort.Slice(versions, func(i, j int) bool {
		if versions[i].Kind == versions[j].Kind {
			return versions[i].Name < versions[j].Name
		}
		return versions[i].Kind < versions[j].Kind
	})
	return versions, nil
}

type remoteRef struct {
	SHA string
	Ref string
}

func defaultBranch(ctx context.Context, cloneURL string) (string, error) {
	out, err := runGit(ctx, "ls-remote", "--symref", cloneURL, "HEAD")
	if err != nil {
		return "", err
	}
	s := bufio.NewScanner(strings.NewReader(out))
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "ref:") && strings.Contains(line, "\tHEAD") {
			parts := strings.SplitN(line, "\t", 2)
			if len(parts) == 2 {
				ref := strings.TrimSpace(strings.TrimPrefix(parts[0], "ref:"))
				if strings.HasPrefix(ref, "refs/heads/") {
					return strings.TrimPrefix(ref, "refs/heads/"), nil
				}
			}
		}
	}
	return "", errors.New("default branch not found")
}

func findTag(ctx context.Context, cloneURL, tag string) (string, error) {
	if tag == "" {
		return "", errors.New("empty tag")
	}
	sha, err := lsRemoteSingle(ctx, cloneURL, "refs/tags/"+tag+"^{}")
	if err == nil {
		return sha, nil
	}
	return lsRemoteSingle(ctx, cloneURL, "refs/tags/"+tag)
}

func lsRemoteSingle(ctx context.Context, cloneURL, ref string) (string, error) {
	entries, err := lsRemote(ctx, cloneURL, ref)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.Ref == ref {
			return e.SHA, nil
		}
	}
	return "", fmt.Errorf("ref not found: %s", ref)
}

func lsRemote(ctx context.Context, cloneURL string, refs ...string) ([]remoteRef, error) {
	args := []string{"ls-remote", "--refs", cloneURL}
	args = append(args, refs...)
	out, err := runGit(ctx, args...)
	if err != nil {
		return nil, err
	}
	items := []remoteRef{}
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 2 {
			continue
		}
		items = append(items, remoteRef{SHA: parts[0], Ref: parts[1]})
	}
	if len(items) == 0 {
		return nil, errors.New("no refs found")
	}
	return items, nil
}

func runGit(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}
