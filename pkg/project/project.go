package project

import (
	"github.com/spf13/viper"
)

type Project struct {
	Dir    string
	config *viper.Viper
}

func (project *Project) GetTemplate() string {
	return project.config.GetString("manala.template")
}
