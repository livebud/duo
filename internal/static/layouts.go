package static

import (
	"io/fs"

	"github.com/livebud/duo/internal/resolver"
)

type layoutResolver struct {
	fsys fs.FS
}

func (r *layoutResolver) Resolve(res *resolver.Resolve) (*resolver.File, error) {
	// TODO: finish me
	return nil, fs.ErrNotExist
}
