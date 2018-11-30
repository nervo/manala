package template

import (
	"errors"
)

var (
	ErrNotFound = errors.New("template not found")
)

type Interface interface {
	GetDir() string
}

func New(dir string) Interface {
	return &template{
		dir: dir,
	}
}

type template struct {
	dir string
}

func (tpl *template) GetDir() string {
	return tpl.dir
}
