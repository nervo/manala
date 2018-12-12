package template

import (
	"github.com/apex/log"
	"github.com/asaskevich/govalidator"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

type FactoryInterface interface {
	Create(name string, fs afero.Fs) (Interface, error)
}

func NewFactory(logger log.Interface) FactoryInterface {
	return &factory{
		logger: logger,
	}
}

type factory struct {
	logger log.Interface
}

func (fa *factory) Create(name string, fs afero.Fs) (Interface, error) {
	vpr := viper.New()
	vpr.SetFs(fs)

	vpr.SetConfigName("manala")
	vpr.AddConfigPath("/")

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
		fs:     fs,
		config: cfg,
	}

	return tpl, nil
}
