package manager

import (
	"manala/pkg/repository"
	"manala/pkg/template"
)

type store struct {
	repositoryFactory repository.FactoryInterface
	repositories      map[string]repository.Interface
}

// Get repository
func (str *store) GetRepository(src string) (repository.Interface, error) {
	// Check if repository already in store
	if rep, ok := str.repositories[src]; ok {
		return rep, nil
	}

	// Create repository
	rep, err := str.repositoryFactory.Create(src)
	if err != nil {
		return nil, err
	}

	// Store repository
	str.repositories[src] = rep

	return rep, nil
}

type Interface interface {
	Walk(fn repository.WalkFunc) error
	Get(name string) (template.Interface, error)
}

func New(repositoryFactory repository.FactoryInterface, repositorySrc string) Interface {
	return &manager{
		store: &store{
			repositoryFactory: repositoryFactory,
			repositories:      make(map[string]repository.Interface),
		},
		repositorySrc: repositorySrc,
	}
}

type manager struct {
	store         *store
	repositorySrc string
}

// Walk into templates
func (mng *manager) Walk(fn repository.WalkFunc) error {
	// Get repository
	rep, err := mng.store.GetRepository(mng.repositorySrc)
	if err != nil {
		return err
	}

	err = rep.Walk(fn)

	return nil
}

// Get template
func (mng *manager) Get(name string) (template.Interface, error) {
	// Get repository
	rep, err := mng.store.GetRepository(mng.repositorySrc)
	if err != nil {
		return nil, err
	}

	// Get template
	tpl, err := rep.Get(name)
	if err != nil {
		return nil, err
	}

	return tpl, nil
}
