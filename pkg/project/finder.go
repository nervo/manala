package project

import (
	"github.com/apex/log"
	"path/filepath"
)

type FinderInterface interface {
	Find(dir string) (*Project, error)
}

type Finder struct {
	Factory FactoryInterface
	Logger  log.Interface
}

// Find a project by browsing dir then its parents up to root
func (finder *Finder) Find(dir string) (*Project, error) {
	// Todo: oh god... this algorithm sucks... how the hell git do ?
	lastDir := ""
	for dir != lastDir {
		lastDir = dir
		finder.Logger.WithField("dir", dir).Debug("Searching project...")
		if p, err := finder.Factory.Create(dir); err == nil {
			return p, nil
		}
		dir = filepath.Dir(dir)
	}

	return nil, ErrNotFound
}
