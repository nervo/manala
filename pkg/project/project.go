package project

import (
	"github.com/spf13/afero"
)

/***********/
/* Project */
/***********/

type Interface interface {
	GetFs() afero.Fs
	GetTemplate() string
	GetRepository() string
	GetOptions() map[string]interface{}
}

type config struct {
	Template   string `mapstructure:"template" valid:"required"`
	Repository string `mapstructure:"repository"`
}

type project struct {
	fs      afero.Fs
	config  config
	options map[string]interface{}
}

func (prj *project) GetFs() afero.Fs {
	return prj.fs
}

func (prj *project) GetTemplate() string {
	return prj.config.Template
}

func (prj *project) GetRepository() string {
	return prj.config.Repository
}

func (prj *project) GetOptions() map[string]interface{} {
	return prj.options
}
