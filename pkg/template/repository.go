package template

import (
	"errors"
	"github.com/spf13/afero"
	"path/filepath"
)

var (
	ErrNotFound = errors.New("template not found")
)

type RepositoryInterface interface {
	GetDir() string
	Get(name string) (Interface, error)
}

type repository struct {
	dir string
	fs  afero.Fs
}

func (r *repository) GetDir() string {
	return r.dir
}

func (r *repository) Get(name string) (Interface, error) {
	dir := filepath.Join(r.GetDir(), name)

	if ok, _ := afero.DirExists(r.fs, dir); !ok {
		return nil, ErrNotFound
	}

	// Instantiate template
	tpl := &template{
		dir: dir,
	}

	return tpl, nil
}
