package project

import (
	"github.com/apex/log"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

type FactoryInterface interface {
	Create(dir string) (Interface, error)
}

func NewFactory(fs afero.Fs, logger log.Interface) FactoryInterface {
	return &factory{
		fs:     fs,
		logger: logger,
	}
}

type factory struct {
	fs     afero.Fs
	logger log.Interface
}

func (fa *factory) Create(dir string) (Interface, error) {
	cfg := viper.New()
	cfg.SetFs(fa.fs)

	cfg.SetConfigName("manala")
	cfg.AddConfigPath(dir)

	fa.logger.WithField("dir", dir).Debug("Reading project config...")

	if err := cfg.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError, viper.UnsupportedConfigError, viper.ConfigParseError:
			return nil, ErrNotFound
		}
		return nil, ErrNotFound
	}

	prj := &project{
		dir:    dir,
		config: cfg,
	}

	if prj.GetTemplate() == "" {
		return nil, ErrTemplateNotDefined
	}

	return prj, nil
}
