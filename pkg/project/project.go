package project

import (
	"errors"
	"github.com/spf13/afero"
)

var (
	ErrNotFound = errors.New("project not found")
	ErrConfig   = errors.New("project config invalid")
)

type Interface interface {
	GetFs() afero.Fs
	GetTemplate() string
}

type config struct {
	Template string `mapstructure:"template" valid:"required"`
}

type project struct {
	fs     afero.Fs
	config config
}

func (prj *project) GetFs() afero.Fs {
	return prj.fs
}

func (prj *project) GetTemplate() string {
	return prj.config.Template
}
