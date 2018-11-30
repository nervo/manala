package repository

import (
	"errors"
	"github.com/apex/log"
	"github.com/spf13/afero"
	"manala/pkg/template"
	"path/filepath"
	"strings"
)

var (
	ErrUnclonable = errors.New("repository unclonable")
	ErrUnopenable = errors.New("repository unopenable")
	ErrInvalid    = errors.New("repository invalid")
)

type Interface interface {
	GetDir() string
	Get(name string) (template.Interface, error)
	Walk(fn WalkFunc) error
}

func New(dir string, fs afero.Fs, templateFactory template.FactoryInterface, logger log.Interface) *repository {
	return &repository{
		dir:             dir,
		fs:              fs,
		templateFactory: templateFactory,
		logger:          logger,
		templates:       make(map[string]template.Interface),
	}
}

type repository struct {
	dir             string
	fs              afero.Fs
	templateFactory template.FactoryInterface
	logger          log.Interface
	templates       map[string]template.Interface
}

func (rep *repository) GetDir() string {
	return rep.dir
}

func (rep *repository) Get(name string) (template.Interface, error) {
	rep.logger.WithField("name", name).Debug("Getting template...")

	// Check if template already in repository
	if tpl, ok := rep.templates[name]; ok {
		rep.logger.WithField("name", name).Debug("Returning template from repository...")
		return tpl, nil
	}

	dir := filepath.Join(rep.GetDir(), name)

	if ok, _ := afero.DirExists(rep.fs, dir); !ok {
		return nil, template.ErrNotFound
	}

	// Instantiate template
	tpl, err := rep.templateFactory.Create(
		name,
		dir,
	)

	return tpl, err
}

type WalkFunc func(tpl template.Interface)

// Walk into templates
func (rep *repository) Walk(fn WalkFunc) error {
	files, err := afero.ReadDir(rep.fs, rep.GetDir())
	if err != nil {
		rep.logger.WithError(err).Fatal("Error walking into templates")
	}

	for _, file := range files {
		// Exclude dot files
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}
		if file.IsDir() {
			tpl, err := rep.Get(file.Name())
			if err != nil {
				return err
			}

			fn(tpl)
		}
	}

	return nil
}
