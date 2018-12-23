package template

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"reflect"
	"strings"
)

/*************/
/* Sync Unit */
/*************/

type SyncUnit struct {
	Source      string `mapstructure:"source"`
	Destination string `mapstructure:"destination"`
	Template    string `mapstructure:"template"`
}

// Returns a DecodeHookFunc that converts strings to syncUnit
func StringToSyncUnitHookFunc() mapstructure.DecodeHookFunc {
	return func(rf reflect.Type, rt reflect.Type, data interface{}) (interface{}, error) {
		if rf.Kind() != reflect.String {
			return data, nil
		}
		if rt != reflect.TypeOf(SyncUnit{}) {
			return data, nil
		}

		src := data.(string)
		dst := src
		tmpl := ""

		// Separate source / destination
		u := strings.Split(src, " ")
		if len(u) > 1 {
			src = u[0]
			dst = u[1]
		}

		// Separate template / source
		v := strings.Split(src, ":")
		if len(v) > 1 {
			tmpl = v[0]
			src = v[1]
			if len(u) < 2 {
				dst = src
			}
		}

		return SyncUnit{
			Source:      src,
			Destination: dst,
			Template:    tmpl,
		}, nil
	}
}

/************/
/* Template */
/************/

type Interface interface {
	GetName() string
	GetFs() afero.Fs
	GetDescription() string
	GetSync() []SyncUnit
}

type config struct {
	Description string     `mapstructure:"description" valid:"required"`
	Sync        []SyncUnit `mapstructure:"sync"`
}

type template struct {
	name   string
	fs     afero.Fs
	config config
}

func (tpl *template) GetName() string {
	return tpl.name
}

func (tpl *template) GetFs() afero.Fs {
	return tpl.fs
}

func (tpl *template) GetDescription() string {
	return tpl.config.Description
}

func (tpl *template) GetSync() []SyncUnit {
	return tpl.config.Sync
}
