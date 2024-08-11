package static

import (
	"io/fs"

	"github.com/livebud/duo/internal/resolver"
)

type errorResolver struct {
	fsys fs.FS
}

func (r *errorResolver) Resolve(res *resolver.Resolve) (*resolver.File, error) {
	// TODO: finish me
	return nil, fs.ErrNotExist
}
