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

func (p *project) GetDir() string {
	return p.dir
}

func (p *project) GetTemplate() string {
	return p.config.GetString("manala.template")
}
