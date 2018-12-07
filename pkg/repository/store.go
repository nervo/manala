package repository

import (
	"github.com/apex/log"
)

type StoreInterface interface {
	Get(src string) (Interface, error)
}

func NewStore(factory FactoryInterface, logger log.Interface) StoreInterface {
	return &store{
		factory:      factory,
		logger:       logger,
		repositories: make(map[string]Interface),
	}
}

type store struct {
	factory      FactoryInterface
	logger       log.Interface
	repositories map[string]Interface
}

func (str *store) Get(src string) (Interface, error) {
	str.logger.WithField("src", src).Debug("Getting repository...")

	// Check if repository already in store
	if r, ok := str.repositories[src]; ok {
		str.logger.WithField("src", src).Debug("Returning repository from store...")
		return r, nil
	}

	// Instantiate repository
	rep, err := str.factory.Create(src)
	if err != nil {
		return nil, err
	}

	// Store repository
	str.repositories[src] = rep

	return rep, nil
}
