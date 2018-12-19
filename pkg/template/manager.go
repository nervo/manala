package template

import (
	"errors"
	"github.com/apex/log"
	"github.com/asaskevich/govalidator"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"manala/pkg/repository"
	"reflect"
	"strings"
)

var (
	ErrNotFound = errors.New("template not found")
	ErrConfig   = errors.New("template config invalid")
)

var supportedConfigNames = []*struct {
	name     string
	required bool
}{
	{".manala", true},
	{".manala.local", false},
}

type ManagerInterface interface {
	Create(name string, fs afero.Fs) (Interface, error)
	Walk(fn ManagerWalkFunc) error
	Get(name string) (Interface, error)
	WithRepositorySrc(src string) ManagerInterface
}

func NewManager(repositoryManager repository.ManagerInterface, logger log.Interface, repositorySrc string) ManagerInterface {
	return &manager{
		managerCore: &managerCore{
			repositoryManager: repositoryManager,
			logger:            logger,
			repositories:      make(map[string]repository.Interface),
			templates:         make(map[string]map[string]Interface),
		},
		repositorySrc: repositorySrc,
	}
}

type managerCore struct {
	repositoryManager repository.ManagerInterface
	logger            log.Interface
	repositories      map[string]repository.Interface
	templates         map[string]map[string]Interface
}

type manager struct {
	*managerCore
	repositorySrc string
}

func (mgr *manager) Create(name string, fs afero.Fs) (Interface, error) {
	vpr := viper.New()
	vpr.SetFs(fs)

	vpr.AddConfigPath("/")

	// Configs
	for _, cfg := range supportedConfigNames {
		vpr.SetConfigName(cfg.name)

		var err error

		if cfg.required {
			err = vpr.ReadInConfig()
		} else {
			err = vpr.MergeInConfig()
		}

		if err != nil {
			switch err.(type) {
			case viper.ConfigFileNotFoundError, viper.UnsupportedConfigError:
				if cfg.required {
					return nil, ErrNotFound
				}
			case viper.ConfigParseError:
				return nil, ErrConfig
			default:
				return nil, err
			}
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

// Get repository
func (mgr *manager) getRepository(src string) (repository.Interface, error) {
	// Check if repository already in store
	if rep, ok := mgr.repositories[src]; ok {
		return rep, nil
	}

	// Create repository
	rep, err := mgr.repositoryManager.Create(src)
	if err != nil {
		// Todo: what about storing "nil" value for template name to speed up next error resolving ?
		return nil, err
	}

	// Store repository
	mgr.repositories[src] = rep

	return rep, nil
}

// Get template
func (mgr *manager) getTemplate(name string, rep repository.Interface) (Interface, error) {

	templates, ok := mgr.templates[rep.GetSrc()]
	if !ok {
		mgr.templates[rep.GetSrc()] = make(map[string]Interface)
		templates = mgr.templates[rep.GetSrc()]
	}

	// Check if template already in store
	if tpl, ok := templates[name]; ok {
		return tpl, nil
	}

	// Todo: is this really usesful ? trying to create the template should be enough...
	if ok, _ := afero.DirExists(rep.GetFs(), name); !ok {
		return nil, ErrNotFound
	}

	// Create template
	tpl, err := mgr.Create(
		name,
		afero.NewBasePathFs(rep.GetFs(), name),
	)
	if err != nil {
		// Todo: what about storing "nil" value for template name to speed up next error resolving ?
		return nil, err
	}

	// Store template
	templates[name] = tpl

	return tpl, nil
}

type ManagerWalkFunc func(tpl Interface)

// Walk into templates
func (mgr *manager) Walk(fn ManagerWalkFunc) error {
	// Get repository
	rep, err := mgr.getRepository(mgr.repositorySrc)
	if err != nil {
		return err
	}

	files, err := afero.ReadDir(rep.GetFs(), "")
	if err != nil {
		mgr.logger.WithError(err).Fatal("Error walking into templates")
	}

	for _, file := range files {
		// Exclude dot files
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}
		if file.IsDir() {
			tpl, err := mgr.getTemplate(file.Name(), rep)
			if err != nil {
				return err
			}

			fn(tpl)
		}
	}

	return nil
}

// Get template
func (mgr *manager) Get(name string) (Interface, error) {
	// Get repository
	rep, err := mgr.getRepository(mgr.repositorySrc)
	if err != nil {
		return nil, err
	}

	return mgr.getTemplate(name, rep)
}

// With repository source
func (mgr *manager) WithRepositorySrc(src string) ManagerInterface {
	return &manager{
		managerCore:   mgr.managerCore,
		repositorySrc: src,
	}
}
