package project

import (
	"errors"
	"github.com/apex/log"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var (
	ErrNotFound           = errors.New("project not found")
	ErrTemplateNotDefined = errors.New("project template not defined")
)

type FactoryInterface interface {
	Create(dir string) (*Project, error)
}

type Factory struct {
	Fs     afero.Fs
	Logger log.Interface
}

func (factory *Factory) Create(dir string) (*Project, error) {
	config := viper.New()
	config.SetFs(factory.Fs)

	config.SetConfigName("manala")
	config.AddConfigPath(dir)

	factory.Logger.WithField("dir", dir).Debug("Reading project config...")

	if err := config.ReadInConfig(); err != nil {
		// Todo: Returns more specifics error, depending on the nature of err
		return nil, ErrNotFound
	}

	project := &Project{
		Dir:    dir,
		config: config,
	}

	if project.GetTemplate() == "" {
		return nil, ErrTemplateNotDefined
	}

	return project, nil
}
