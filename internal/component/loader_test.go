package component

import "testing"

func TestNormalizeAliases(t *testing.T) {
	c, err := Normalize(RawComponent{Name: "A", Source: "github://o/r", Ver: "1.0.0", Dir: "dist"}, "a.json")
	if err != nil {
		t.Fatal(err)
	}
	if c.Version != "1.0.0" || c.Directory != "dist" {
		t.Fatalf("unexpected normalize: %#v", c)
	}
}

func TestNormalizeConflicts(t *testing.T) {
	_, err := Normalize(RawComponent{Name: "A", Source: "github://o/r", Ver: "1", Version: "2"}, "a.json")
	if err == nil {
		t.Fatal("expected conflict error")
	}
}
