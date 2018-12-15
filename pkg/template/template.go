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
	GetSync() []syncUnit
}

type config struct {
	Description string     `mapstructure:"description" valid:"required"`
	Sync        []syncUnit `mapstructure:"sync"`
}

type syncUnit struct {
	Source      string `mapstructure:"source"`
	Destination string `mapstructure:"destination"`
	Template    string `mapstructure:"template"`
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

func (tpl *template) GetSync() []syncUnit {
	return tpl.config.Sync
}
