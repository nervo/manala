package template

import (
	"errors"
	"github.com/apex/log"
	"github.com/asaskevich/govalidator"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"manala/pkg/repository"
	"path"
	"strings"
)

/**********/
/* Errors */
/**********/

var (
	ErrNotFound = errors.New("template not found")
	ErrConfig   = errors.New("template config invalid")
)

/**********/
/* Config */
/**********/

var supportedConfigNames = []*struct {
	name     string
	required bool
}{
	{".manala", true},
	{".manala.local", false},
}

/********************/
/* Managed Template */
/********************/

type ManagedTemplate struct {
	Interface
	dir string
}

func (tmpl *ManagedTemplate) GetDir() string {
	return tmpl.dir
}

/***********/
/* Manager */
/***********/

type ManagerInterface interface {
	Create(name string, fs afero.Fs) (*template, error)
	Walk(fn ManagerWalkFunc) error
	Get(name string) (*ManagedTemplate, error)
	WithRepositorySrc(src string) ManagerInterface
}

type manager struct {
	repositoryManager repository.ManagerInterface
	logger            log.Interface
	repositories      map[string]*repository.ManagedRepository
	templates         map[string]map[string]*ManagedTemplate
}

func NewSingleRepositoryManager(repositoryManager repository.ManagerInterface, logger log.Interface, repositorySrc string) *singleRepositoryManager {
	return &singleRepositoryManager{
		manager: &manager{
			repositoryManager: repositoryManager,
			logger:            logger,
			repositories:      make(map[string]*repository.ManagedRepository),
			templates:         make(map[string]map[string]*ManagedTemplate),
		},
		repositorySrc: repositorySrc,
	}
}

type singleRepositoryManager struct {
	*manager
	repositorySrc string
}

func (mgr *singleRepositoryManager) Create(name string, fs afero.Fs) (*template, error) {
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
			StringToSyncUnitHookFunc(),
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
	return &template{
		name:   name,
		fs:     fs,
		config: cfg,
	}, nil
}

// Get repository
func (mgr *singleRepositoryManager) getRepository(src string) (*repository.ManagedRepository, error) {
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
func (mgr *singleRepositoryManager) getTemplate(name string, rep *repository.ManagedRepository) (*ManagedTemplate, error) {

	templates, ok := mgr.templates[rep.GetSrc()]
	if !ok {
		mgr.templates[rep.GetSrc()] = make(map[string]*ManagedTemplate)
		templates = mgr.templates[rep.GetSrc()]
	}

	// Check if template already in store
	if tmpl, ok := templates[name]; ok {
		return tmpl, nil
	}

	// Todo: is this really usesful ? trying to create the template should be enough...
	if ok, _ := afero.DirExists(rep.GetFs(), name); !ok {
		return nil, ErrNotFound
	}

	// Create template
	tmpl, err := mgr.Create(
		name,
		afero.NewBasePathFs(rep.GetFs(), name),
	)
	if err != nil {
		// Todo: what about storing "nil" value for template name to speed up next error resolving ?
		return nil, err
	}

	mgrTmpl := &ManagedTemplate{
		Interface: tmpl,
		dir:       path.Join(rep.GetDir(), name),
	}

	// Store template
	templates[name] = mgrTmpl

	return mgrTmpl, nil
}

type ManagerWalkFunc func(tmpl *ManagedTemplate)

// Walk into templates
func (mgr *singleRepositoryManager) Walk(fn ManagerWalkFunc) error {
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
			tmpl, err := mgr.getTemplate(file.Name(), rep)
			if err != nil {
				return err
			}

			fn(tmpl)
		}
	}

	return nil
}

// Get template
func (mgr *singleRepositoryManager) Get(name string) (*ManagedTemplate, error) {
	// Get repository
	repo, err := mgr.getRepository(mgr.repositorySrc)
	if err != nil {
		return nil, err
	}

	return mgr.getTemplate(name, repo)
}

// With repository source
func (mgr *singleRepositoryManager) WithRepositorySrc(src string) ManagerInterface {
	return &singleRepositoryManager{
		manager:       mgr.manager,
		repositorySrc: src,
	}
}
