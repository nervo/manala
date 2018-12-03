package template

import (
	"errors"
	"github.com/spf13/afero"
)

var (
	ErrNotFound = errors.New("template not found")
	ErrConfig   = errors.New("template config invalid")
)

type Interface interface {
	GetName() string
	GetFs() afero.Fs
	GetDescription() string
	GetSync() []string
}

type config struct {
	Description string   `mapstructure:"description" valid:"required"`
	Sync        []string `mapstructure:"sync"`
}

type template struct {
	name   string
	fs     afero.Fs
	config config
}

func (tpl *template) GetName() string {
	return tpl.name
}

func (tpl *template) GetFs() afero.Fs {
	return tpl.fs
}

func (tpl *template) GetDescription() string {
	return tpl.config.Description
}

func (tpl *template) GetSync() []string {
	return tpl.config.Sync
}
