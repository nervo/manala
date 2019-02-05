package zroject

import (
	"github.com/asaskevich/govalidator"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

/**********/
/* Config */
/**********/

type config struct {
	Template   string                 `valid:"manala_template_name~template invalid,required~template not found" yaml:"template"`
	Repository string                 `valid:"manala_repository_description~repository invalid"                         yaml:"repository,omitempty"`
	Options    map[string]interface{} `valid:""                                                      yaml:"options,omitempty"`
}

func (c *config) GetTemplateName() string {
	return c.Template
}

func (c *config) GetRepositorySource() string {
	return c.Repository
}

func (c *config) GetOptions() map[string]interface{} {
	return c.Options
}

/**********/
/* Errors */
/**********/

type ErrorConfigNotFound struct{ error }
type ErrorConfigMalformed struct{ error }

//type ErrorConfigMerge struct{ error }
type ErrorConfigInvalid struct{ error }

/*********/
/* Event */
/*********/

/*
type configHandlerEvent struct {
	Path string
}
*/

/***********/
/* Handler */
/***********/

/*
type ConfigLoaderInterface interface {
	Load() (*config, error)
}
*/

/*
type ConfigDumperInterface interface {
	Dump(c *config) error
}
*/

/*
type yamlConfigHandler struct {
	path   string
	Events chan configHandlerEvent
}
*/

/*
func NewYamlConfigHandler(path string) *yamlConfigHandler {
	return &yamlConfigHandler{
		path:   path,
		Events: make(chan configHandlerEvent),
	}
}
*/

func LoadConfig(file string) (*config, error) {
	// Read content
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, &ErrorConfigNotFound{err}
	}

	var c config

	// Unmarshal content into config
	if err := yaml.Unmarshal(content, &c); err != nil {
		return nil, &ErrorConfigMalformed{err}
	}

	// Validation
	_, err = govalidator.ValidateStruct(c)
	if err != nil {
		return &c, &ErrorConfigInvalid{err}
	}

	return &c, nil
}

/*
func MergeConfig(configs ...*config) (*config, error) {
	var c config

	for _, config := range configs {
		err := mergo.Merge(&c, config, mergo.WithOverride)
		if err != nil {
			return nil, &ErrorConfigMerge{err}
		}
	}

	// Validation
	_, err := govalidator.ValidateStruct(c)
	if err != nil {
		return &c, &ErrorConfigInvalid{err}
	}

	return &c, nil
}
*/

/*
func (h *yamlConfigHandler) Dump(c *config) error {
	// Marshal config into content
	content, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	// Write content
	err = afero.WriteFile(afero.NewOsFs(), h.path, content, 0666)
	if err != nil {
		return err
	}

	return nil
}
*/
