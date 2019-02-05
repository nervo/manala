package zemplate

import (
	"github.com/spf13/afero"
)

const configFile = ".manala.yaml"

/************/
/* Template */
/************/

type Interface interface {
	GetFs() afero.Fs
	GetName() string
	GetDescription() string
	GetOptions() map[string]interface{}
	//GetSync() []SyncUnit
}

type template struct {
	dir     string
	options map[string]interface{}
}

func (t *template) GetFs() afero.Fs {
	// Todo: implement some kind of on-demand cache
	return afero.NewBasePathFs(
		afero.NewOsFs(),
		t.dir,
	)
}

func (t *template) GetOptions() map[string]interface{} {
	return t.options
}

/*************/
/* Reference */
/*************/

type ReferenceInterface interface {
	GetTemplateName() string
}
