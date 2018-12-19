package template

import (
	"errors"
	"github.com/apex/log"
	"github.com/asaskevich/govalidator"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"reflect"
	"strings"
)

var (
	ErrNotFound = errors.New("template not found")
	ErrConfig   = errors.New("template config invalid")
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

	if vpr = vpr.Sub("manala"); vpr == nil {
		return nil, ErrConfig
	}

	var cfg config

	// Unmarshalling
	err := vpr.Unmarshal(&cfg, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			stringToSyncUnitHookFunc(),
		),
	))
	if err != nil {
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

// Returns a DecodeHookFunc that converts strings to syncUnit
func stringToSyncUnitHookFunc() mapstructure.DecodeHookFunc {
	return func(rf reflect.Type, rt reflect.Type, data interface{}) (interface{}, error) {
		if rf.Kind() != reflect.String {
			return data, nil
		}
		if rt != reflect.TypeOf(syncUnit{}) {
			return data, nil
		}

		src := data.(string)
		dst := src
		tpl := ""

		// Separate source / destination
		u := strings.Split(src, " ")
		if len(u) > 1 {
			src = u[0]
			dst = u[1]
		}

		// Separate template / source
		v := strings.Split(src, ":")
		if len(v) > 1 {
			tpl = v[0]
			src = v[1]
			if len(u) < 2 {
				dst = src
			}
		}

		return syncUnit{
			Source:      src,
			Destination: dst,
			Template:    tpl,
		}, nil
	}
}
