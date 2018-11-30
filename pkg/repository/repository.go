package repository

import (
	"errors"
	"github.com/spf13/afero"
	"manala/pkg/template"
	"path/filepath"
)

var (
	ErrUnclonable = errors.New("repository unclonable")
	ErrUnopenable = errors.New("repository unopenable")
	ErrInvalid    = errors.New("repository invalid")
)

type Interface interface {
	GetDir() string
	Get(name string) (template.Interface, error)
}

type repository struct {
	dir string
	fs  afero.Fs
}

func (rep *repository) GetDir() string {
	return rep.dir
}

func (rep *repository) Get(name string) (template.Interface, error) {
	dir := filepath.Join(rep.GetDir(), name)

	if ok, _ := afero.DirExists(rep.fs, dir); !ok {
		return nil, template.ErrNotFound
	}

	// Instantiate template
	tpl := template.New(
		dir,
	)

	return tpl, nil
}
