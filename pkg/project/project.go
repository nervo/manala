package project

import (
	"errors"
	"github.com/spf13/viper"
)

var (
	ErrNotFound           = errors.New("project not found")
	ErrTemplateNotDefined = errors.New("project template not defined")
)

type Project struct {
	Dir    string
	config *viper.Viper
}

func (project *Project) GetTemplate() string {
	return project.config.GetString("manala.template")
}

func New(dir string) (*Project, error) {
	config := viper.New()

	config.SetConfigName("manala")
	config.AddConfigPath(dir)

	if err := config.ReadInConfig(); err != nil {
		// Todo: Returns more specifics error, depending on the nature of err
		return nil, ErrNotFound
	}

	p := &Project{
		Dir:    dir,
		config: config,
	}

	if p.GetTemplate() == "" {
		return nil, ErrTemplateNotDefined
	}

	return p, nil
}
