package source

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var schemeRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*$`)

func Parse(input string) (*SourceURI, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return nil, fmt.Errorf("source is empty")
	}
	for _, r := range raw {
		if unicode.IsControl(r) {
			return nil, fmt.Errorf("source contains control characters")
		}
	}
	idx := strings.Index(raw, "://")
	if idx <= 0 {
		return nil, fmt.Errorf("source must include a valid scheme and ://")
	}
	scheme := strings.ToLower(raw[:idx])
	if !schemeRe.MatchString(scheme) {
		return nil, fmt.Errorf("invalid source scheme %q", scheme)
	}
	target := raw[idx+3:]
	if target == "" {
		return nil, fmt.Errorf("source target is empty")
	}
	return &SourceURI{Scheme: scheme, Target: target}, nil
}
