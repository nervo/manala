package project

import (
	"github.com/apex/log"
	"path/filepath"
)

type FinderInterface interface {
	Find(dir string) (Interface, error)
}

func NewFinder(factory FactoryInterface, logger log.Interface) FinderInterface {
	return &finder{
		factory: factory,
		logger:  logger,
	}
}

type finder struct {
	factory FactoryInterface
	logger  log.Interface
}

// Find a project by browsing dir then its parents up to root
func (fi *finder) Find(dir string) (Interface, error) {
	// Todo: oh god... this algorithm sucks... how the hell git do ?
	lastDir := ""
	for dir != lastDir {
		lastDir = dir
		fi.logger.WithField("dir", dir).Debug("Searching project...")
		p, err := fi.factory.Create(dir)
		if err == nil {
			return p, nil
		}
		dir = filepath.Dir(dir)
	}

	return nil, ErrNotFound
}
