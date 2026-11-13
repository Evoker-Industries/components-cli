package lockfile

import (
	"encoding/json"
	"os"
	"time"

	"github.com/Evoker-Industries/components-cli/internal/util"
)

func Load(path string) (*File, error) {
	abs, err := util.MustAbs(path)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return &File{Schema: "component-lock/v1", GeneratedAt: time.Now().UTC(), Components: map[string]ComponentLock{}}, nil
		}
		return nil, err
	}
	var v FileView
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	genAt, _ := time.Parse(time.RFC3339, v.GeneratedAt)
	if genAt.IsZero() {
		genAt = time.Now().UTC()
	}
	if v.Components == nil {
		v.Components = map[string]ComponentLock{}
	}
	if v.Schema == "" {
		v.Schema = "component-lock/v1"
	}
	return &File{Schema: v.Schema, GeneratedAt: genAt, Components: v.Components}, nil
}
