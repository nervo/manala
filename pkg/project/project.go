package project

import (
	"github.com/spf13/viper"
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
