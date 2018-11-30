package project

import (
	"errors"
)

var (
	ErrNotFound = errors.New("project not found")
	ErrConfig   = errors.New("project config invalid")
)

type Interface interface {
	GetTemplate() string
	GetDir() string
}

type config struct {
	Template string `mapstructure:"template" valid:"required"`
}

type project struct {
	config config
	dir    string
}

func (prj *project) GetTemplate() string {
	return prj.config.Template
}

func (prj *project) GetDir() string {
	return prj.dir
}
