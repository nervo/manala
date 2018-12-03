package project

import (
	"github.com/apex/log"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
)

type FinderInterface interface {
	Find(dir string) (Interface, error)
	Walk(dir string, fn WalkFunc) error
}

func NewFinder(fs afero.Fs, factory FactoryInterface, logger log.Interface) FinderInterface {
	return &finder{
		fs:      fs,
		factory: factory,
		logger:  logger,
	}
}

type finder struct {
	fs      afero.Fs
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
		p, err := fi.factory.Create(afero.NewBasePathFs(fi.fs, dir))
		if err == nil {
			return p, nil
		}
		dir = filepath.Dir(dir)
	}

	return nil, ErrNotFound
}

type WalkFunc func(project Interface)

// Find projects recursively starting from dir
func (fi *finder) Walk(dir string, fn WalkFunc) error {

	err := afero.Walk(fi.fs, dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		fi.logger.WithField("path", path).Debug("Searching project...")
		p, err := fi.factory.Create(afero.NewBasePathFs(fi.fs, path))
		if err != nil {
			return nil
		}

		fn(p)

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
