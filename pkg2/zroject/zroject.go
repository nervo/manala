package zroject

import (
	"fmt"
	"github.com/imdario/mergo"
	"github.com/spf13/afero"
	"manala/pkg2/zepository"
	"path/filepath"
	"strings"
)

const configFile = ".manala.yaml"

/***********/
/* Project */
/***********/

type Interface interface {
	GetFs() afero.Fs
	GetOptions(scopes ...string) (map[string]interface{}, error)
}

type project struct {
	dir    string
	scopes map[string]ScopeInterface
}

/*
func (p *project) addScope(name string, scope ScopeInterface) {
	// Initialize scopes if necessary
	if p.scopes == nil {
		p.scopes = make(map[string]ScopeInterface)
	}

	p.scopes[name] = scope
}
*/

func (p *project) GetFs() afero.Fs {
	// Todo: implement some kind of on-demand cache
	return afero.NewBasePathFs(
		afero.NewOsFs(),
		p.dir,
	)
}

func (p *project) GetOptions(scopes ...string) (map[string]interface{}, error) {
	// Todo: implement some kind of on-demand cache
	options := make(map[string]interface{})

	if len(scopes) == 0 {
		for _, s := range p.scopes {
			_ = mergo.Merge(&options, s.GetOptions(), mergo.WithOverride)
		}
		return options, nil
	} else {
		for _, scope := range scopes {
			s, ok := p.scopes[scope]
			if !ok {
				return nil, &ErrorScopeNotFound{scope}
			}
			_ = mergo.Merge(&options, s.GetOptions(), mergo.WithOverride)
		}
	}

	return options, nil

	/*
		// Single scope
		if len(scopes) == 1 {
			s, ok := p.scopes[scopes[0]]
			if !ok {
				return nil, &ErrorScopeNotFound{scopes[0]}
			}

			return s.GetOptions(), nil
		}

		// Multiple scopes
		options := make(map[string]interface{})
		for _, scope := range scopes {
			s, ok := p.scopes[scope]
			if !ok {
				return nil, &ErrorScopeNotFound{scope}
			}
			_ = mergo.Merge(&options, s.GetOptions(), mergo.WithOverride)
		}

		return options, nil
	*/
}

/*********/
/* Scope */
/*********/

type ScopeInterface interface {
	GetOptions() map[string]interface{}
}

/*
type scope struct {
	name string
	ScopeInterface
}
*/

/**********/
/* Errors */
/**********/

type ErrorScopeNotFound struct{ name string }

func (e ErrorScopeNotFound) Error() string { return fmt.Sprintf("scope %s not found", e.name) }

/***********/
/* Handler */
/***********/

/*
type dirProjectHandler struct {
	dir string
}
*/

/*
func NewDirProjectHandler(dir string) *dirProjectHandler {
	return &dirProjectHandler{
		dir: dir,
	}
}
*/

//func (h *dirProjectHandler) Load() (*project, error) {
func Load(dir string) (*project, error) {
	// Load dist config
	distConfig, err := LoadConfig(
		filepath.Join(dir, configFile),
	)

	if err != nil {
		return nil, err
	}

	// Load local config
	localConfig, err := LoadConfig(
		filepath.Join(dir, strings.TrimSuffix(configFile, filepath.Ext(configFile))+".local"+filepath.Ext(configFile)),
	)

	// Only local malformed errors are relevant
	switch err.(type) {
	case *ErrorConfigMalformed:
		return nil, err
	}

	// Merge configs
	/*
		config, err := MergeConfig(distConfig, localConfig)
		if err != nil {
			return nil, err
		}
	*/

	// Create project
	p := &project{
		dir: dir,
		//options: config.Options,
		scopes: map[string]ScopeInterface{
			"dist":  distConfig,
			"local": localConfig,
		},
	}

	/*
		p.addScope("dist", distConfig)
		p.addScope("local", localConfig)
	*/

	repositoryManager := zepository.NewManager(
		distConfig,
		localConfig,
	)

	fmt.Printf("%#v\n", repositoryManager)

	repositoryManager.LoadTemplate(
		distConfig,
		localConfig,
	)

	return p, nil
}
