package project

import (
	"github.com/apex/log"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
)

type ManagerInterface interface {
	Find(dir string) (Interface, error)
	Walk(dir string, fn WalkFunc) error
}

func NewManager(fs afero.Fs, factory FactoryInterface, logger log.Interface) ManagerInterface {
	return &manager{
		fs:      fs,
		factory: factory,
		logger:  logger,
	}
}

type manager struct {
	fs      afero.Fs
	factory FactoryInterface
	logger  log.Interface
}

// Find a project by browsing dir then its parents up to root
func (mgr *manager) Find(dir string) (Interface, error) {
	// Todo: oh god... this algorithm sucks... how the hell git do ?
	lastDir := ""
	for dir != lastDir {
		lastDir = dir
		mgr.logger.WithField("dir", dir).Debug("Searching project...")
		p, err := mgr.factory.Create(afero.NewBasePathFs(mgr.fs, dir))
		if err == nil {
			return p, nil
		}
		dir = filepath.Dir(dir)
	}

	return nil, ErrNotFound
}

type WalkFunc func(project Interface)

// Find projects recursively starting from dir
func (mgr *manager) Walk(dir string, fn WalkFunc) error {

	err := afero.Walk(mgr.fs, dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		mgr.logger.WithField("path", path).Debug("Searching project...")
		p, err := mgr.factory.Create(afero.NewBasePathFs(mgr.fs, path))
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
