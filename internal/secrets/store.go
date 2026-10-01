package secrets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Evoker-Industries/components-cli/internal/util"
)

type Store struct {
	Path string
	Data map[string]map[string]string
}

func Load(path string) (*Store, error) {
	abs, err := util.MustAbs(path)
	if err != nil {
		return nil, err
	}
	s := &Store{Path: abs, Data: map[string]map[string]string{}}
	b, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(b, &s.Data); err != nil {
		return nil, fmt.Errorf("%s: invalid JSON: %w", abs, err)
	}
	return s, nil
}

func Init(path string, force bool) error {
	abs, err := util.MustAbs(path)
	if err != nil {
		return err
	}
	if _, err := os.Stat(abs); err == nil && !force {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return util.WriteJSONAtomic(abs, map[string]any{}, 0o600)
}

func ensureSafePerm(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if fi.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("secrets file has insecure permissions: %s", path)
	}
	return nil
}

func (s *Store) Save() error {
	if err := ensureSafePerm(s.Path); err != nil {
		return err
	}
	return util.WriteJSONAtomic(s.Path, s.Data, 0o600)
}

func (s *Store) Set(key string, value string) error {
	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return fmt.Errorf("secret key must be in alias.field format")
	}
	if s.Data[parts[0]] == nil {
		s.Data[parts[0]] = map[string]string{}
	}
	s.Data[parts[0]][parts[1]] = value
	return nil
}

func (s *Store) Remove(key string) error {
	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("secret key must be in alias.field format")
	}
	fields := s.Data[parts[0]]
	if fields == nil {
		return nil
	}
	delete(fields, parts[1])
	if len(fields) == 0 {
		delete(s.Data, parts[0])
	}
	return nil
}

func (s *Store) Keys() []string {
	keys := []string{}
	for alias, fields := range s.Data {
		for field := range fields {
			keys = append(keys, alias+"."+field)
		}
	}
	sort.Strings(keys)
	return keys
}
