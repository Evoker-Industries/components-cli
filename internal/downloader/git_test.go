package downloader

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func makeRepo(t *testing.T) string {
	d := t.TempDir()
	run(t, d, "git", "init")
	run(t, d, "git", "config", "user.email", "test@example.com")
	run(t, d, "git", "config", "user.name", "test")
	if err := os.WriteFile(filepath.Join(d, "a.txt"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, d, "git", "add", ".")
	run(t, d, "git", "commit", "-m", "init")
	run(t, d, "git", "tag", "v1.0.0")
	run(t, d, "git", "checkout", "-b", "dev")
	if err := os.WriteFile(filepath.Join(d, "b.txt"), []byte("b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, d, "git", "add", ".")
	run(t, d, "git", "commit", "-m", "dev")
	run(t, d, "git", "checkout", "master")
	return d
}

func TestResolveGitVersion(t *testing.T) {
	repo := makeRepo(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r := GitResolver{}
	res, sha, err := r.Resolve(ctx, repo, "tag:v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if res != "tag:v1.0.0" || sha == "" {
		t.Fatalf("bad resolve for tag: %s %s", res, sha)
	}
	res, sha, err = r.Resolve(ctx, repo, "branch:dev")
	if err != nil {
		t.Fatal(err)
	}
	if res != "branch:dev" || sha == "" {
		t.Fatalf("bad resolve for branch: %s %s", res, sha)
	}
	res, sha, err = r.Resolve(ctx, repo, "latest")
	if err != nil {
		t.Fatal(err)
	}
	if res != "latest" || sha == "" {
		t.Fatalf("bad resolve for latest: %s %s", res, sha)
	}
}

func TestListGitVersions(t *testing.T) {
	repo := makeRepo(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	versions, err := (GitResolver{}).List(ctx, repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) == 0 {
		t.Fatal("expected versions")
	}
}
