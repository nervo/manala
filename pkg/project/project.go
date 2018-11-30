package project

import (
	"errors"
	"github.com/spf13/viper"
)

var (
	ErrNotFound           = errors.New("project not found")
	ErrTemplateNotDefined = errors.New("project template not defined")
)

type Interface interface {
	GetDir() string
	GetTemplate() string
}

type project struct {
	dir    string
	config *viper.Viper
}

func (prj *project) GetDir() string {
	return prj.dir
}

func (prj *project) GetTemplate() string {
	return prj.config.GetString("manala.template")
}
