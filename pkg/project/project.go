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

type Config struct {
	Template   string `mapstructure:"template" valid:"required" yaml:"template"`
	Repository string `mapstructure:"repository" yaml:"repository,omitempty"`
}

type project struct {
	fs      afero.Fs
	config  Config
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
