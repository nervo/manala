package project

import (
	"errors"
	"github.com/apex/log"
	"github.com/asaskevich/govalidator"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var (
	ErrNotFound = errors.New("project not found")
	ErrConfig   = errors.New("project config invalid")
)

type FactoryInterface interface {
	Create(fs afero.Fs) (Interface, error)
}

func NewFactory(logger log.Interface) FactoryInterface {
	return &factory{
		logger: logger,
	}
}

type factory struct {
	logger log.Interface
}

func (fa *factory) Create(fs afero.Fs) (Interface, error) {
	vpr := viper.New()
	vpr.SetFs(fs)

	vpr.AddConfigPath("/")

	// Main config
	vpr.SetConfigName(".manala")

	if err := vpr.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError, viper.UnsupportedConfigError:
			return nil, ErrNotFound
		case viper.ConfigParseError:
			return nil, ErrConfig
		default:
			return nil, err
		}
	}

	// Local config (optional)
	vpr.SetConfigName(".manala.local")

	if err := vpr.MergeInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError, viper.UnsupportedConfigError:
			// Do nothing, as local config is optional
		case viper.ConfigParseError:
			return nil, ErrConfig
		default:
			return nil, err
		}
	}

	// Options
	prjOptions := vpr.AllSettings()

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

	prj := &project{
		fs:      fs,
		config:  cfg,
		options: prjOptions,
	}

	return prj, nil
}
