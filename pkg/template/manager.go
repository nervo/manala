package template

import (
	"github.com/apex/log"
	"github.com/spf13/afero"
	"manala/pkg/repository"
	"strings"
)

type ManagerInterface interface {
	Walk(fn ManagerWalkFunc) error
	Get(name string) (Interface, error)
	WithRepositorySrc(src string) ManagerInterface
}

func NewManager(repositoryFactory repository.FactoryInterface, templateFactory FactoryInterface, logger log.Interface, repositorySrc string) ManagerInterface {
	return &manager{
		managerCore: &managerCore{
			repositoryFactory: repositoryFactory,
			templateFactory:   templateFactory,
			logger:            logger,
			repositories:      make(map[string]repository.Interface),
			templates:         make(map[string]Interface),
		},
		repositorySrc: repositorySrc,
	}
}

type managerCore struct {
	repositoryFactory repository.FactoryInterface
	templateFactory   FactoryInterface
	logger            log.Interface
	repositories      map[string]repository.Interface
	templates         map[string]Interface
}

type manager struct {
	*managerCore
	repositorySrc string
}

// Get repository
func (mgr *manager) getRepository(src string) (repository.Interface, error) {
	// Check if repository already in store
	if rep, ok := mgr.repositories[src]; ok {
		return rep, nil
	}

	// Create repository
	rep, err := mgr.repositoryFactory.Create(src)
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
	// Check if template already in store
	if tpl, ok := mgr.templates[name]; ok {
		return tpl, nil
	}

	// Todo: is this really usesful ? trying to create the template with the factory should be enough...
	if ok, _ := afero.DirExists(rep.GetFs(), name); !ok {
		return nil, ErrNotFound
	}

	// Create template
	tpl, err := mgr.templateFactory.Create(
		name,
		afero.NewBasePathFs(rep.GetFs(), name),
	)
	if err != nil {
		// Todo: what about storing "nil" value for template name to speed up next error resolving ?
		return nil, err
	}

	// Store template
	mgr.templates[name] = tpl

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
