package lockfile

import (
	"time"

	"github.com/Evoker-Industries/components-cli/internal/util"
)

func Write(path string, f *File) error {
	f.GeneratedAt = time.Now().UTC()
	view := FileView{Schema: f.Schema, GeneratedAt: f.GeneratedAt.Format(time.RFC3339), Components: f.Components}
	return util.WriteJSONAtomic(path, view, 0o644)
}
