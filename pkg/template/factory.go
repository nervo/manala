package template

import (
	"github.com/apex/log"
	"github.com/asaskevich/govalidator"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

type FactoryInterface interface {
	Create(name string, dir string) (Interface, error)
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

func (fa *factory) Create(name string, dir string) (Interface, error) {
	vpr := viper.New()
	vpr.SetFs(fa.fs)

	vpr.SetConfigName("manala")
	vpr.AddConfigPath(dir)

	fa.logger.WithField("dir", dir).Debug("Reading template config...")

	if err := vpr.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError, viper.UnsupportedConfigError:
			return nil, ErrNotFound
		case viper.ConfigParseError:
			return nil, ErrConfig
		}
		return nil, ErrNotFound
	}

	if vpr = vpr.Sub("manala"); vpr == nil {
		return nil, ErrConfig
	}

	var cfg config

	// Unmarshalling
	if err := vpr.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Validation
	if _, err := govalidator.ValidateStruct(cfg); err != nil {
		return nil, err
	}

	// Instantiate template
	tpl := &template{
		name:   name,
		config: cfg,
		dir:    dir,
	}

	return tpl, nil
}
