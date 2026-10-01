package source

import "testing"

func TestParseSource(t *testing.T) {
	u, err := Parse("GitHub://owner/repo")
	if err != nil {
		t.Fatal(err)
	}
	if u.Scheme != "github" || u.Target != "owner/repo" {
		t.Fatalf("unexpected parse: %#v", u)
	}
}

func TestParseSourceInvalid(t *testing.T) {
	if _, err := Parse("missing"); err == nil {
		t.Fatal("expected error")
	}
}
