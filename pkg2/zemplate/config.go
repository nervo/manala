package zemplate

import (
	"github.com/asaskevich/govalidator"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func init() {
	govalidator.TagMap["manala_template_name"] = govalidator.Validator(func(str string) bool {
		return govalidator.Matches(str, "^[a-z.-]{1,64}$")
	})
	govalidator.TagMap["manala_template_description"] = govalidator.Validator(func(str string) bool {
		return govalidator.Matches(str, "^[a-z.-]{1,64}$")
	})
}

/**********/
/* Config */
/**********/

type config struct {
	Description string                 `valid:"manala_template_description~description invalid,required~description not found" yaml:"description"`
	Options     map[string]interface{} `valid:""                                                      yaml:"options,omitempty"`
	//Sync        []SyncUnit   `valid:""                                                      yaml:"options,omitempty"`
}

/**********/
/* Errors */
/**********/

type ErrorConfigNotFound struct{ error }
type ErrorConfigMalformed struct{ error }
type ErrorConfigInvalid struct{ error }

/********/
/* Load */
/********/

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
