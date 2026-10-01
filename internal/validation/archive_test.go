package validation

import "testing"

func TestIsSafeExtractPath(t *testing.T) {
	base := t.TempDir()
	cases := map[string]bool{
		"dist/index.js":  true,
		"nested/a/b.txt": true,
		"../etc/passwd":  false,
		"/absolute/path": false,
		"..\\windows\\x": false,
		"C:\\Windows\\x": false,
		"":               false,
		"./":             false,
	}
	for p, want := range cases {
		if got := IsSafeExtractPath(base, p); got != want {
			t.Fatalf("path %q got %v want %v", p, got, want)
		}
	}
}
