package template

import (
	"errors"
)

var (
	ErrNotFound = errors.New("template not found")
	ErrConfig   = errors.New("template config invalid")
)

type Interface interface {
	GetName() string
	GetDescription() string
	GetDir() string
}

type config struct {
	Description string `mapstructure:"description" valid:"required"`
}

type template struct {
	name   string
	config config
	dir    string
}

func (tpl *template) GetName() string {
	return tpl.name
}

func (tpl *template) GetDescription() string {
	return tpl.config.Description
}

func (tpl *template) GetDir() string {
	return tpl.dir
}
