package project

import (
	"errors"
	"github.com/apex/log"
	"github.com/asaskevich/govalidator"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"manala/pkg/syncer"
	"manala/pkg/template"
	"os"
	"path/filepath"
)

var (
	ErrNotFound = errors.New("project not found")
	ErrConfig   = errors.New("project config invalid")
)

var supportedConfigNames = []*struct {
	name     string
	required bool
}{
	{".manala", true},
	{".manala.local", false},
}

func GetSupportedConfigFiles() []string {
	var files []string
	for _, cfg := range supportedConfigNames {
		for _, ext := range viper.SupportedExts {
			files = append(files, cfg.name+"."+ext)
		}
	}

	return files
}

type ManagerInterface interface {
	Create(fs afero.Fs) (Interface, error)
	Find(dir string) (Interface, error)
	Walk(dir string, fn WalkFunc) error
	Sync(prj Interface, tmplMgr template.ManagerInterface, snc syncer.Interface) error
}

func NewManager(fs afero.Fs, logger log.Interface) ManagerInterface {
	return &manager{
		fs:     fs,
		logger: logger,
	}
}

type manager struct {
	fs     afero.Fs
	logger log.Interface
}

func (mgr *manager) Create(fs afero.Fs) (Interface, error) {
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

	// Options
	prjOptions := vpr.AllSettings()

	if vpr = vpr.Sub("manala"); vpr == nil {
		return nil, ErrConfig
	}

	var cfg config

	// Unmarshalling
	if err := vpr.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Validation
	if _, err := govalidator.ValidateStruct(cfg); err != nil {
		return nil, err
	}

	prj := &project{
		fs:      fs,
		config:  cfg,
		options: prjOptions,
	}

	return prj, nil
}

// Find a project by browsing dir then its parents up to root
func (mgr *manager) Find(dir string) (Interface, error) {
	// Todo: oh god... this algorithm sucks... how the hell git do ?
	lastDir := ""
	for dir != lastDir {
		lastDir = dir
		mgr.logger.WithField("dir", dir).Debug("Searching project...")
		p, err := mgr.Create(afero.NewBasePathFs(mgr.fs, dir))
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

		mgr.logger.WithField("dir", path).Debug("Searching project...")
		p, err := mgr.Create(afero.NewBasePathFs(mgr.fs, path))
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

func (mgr *manager) Sync(prj Interface, tmplMgr template.ManagerInterface, snc syncer.Interface) error {
	// Custom project repository
	if prj.GetRepository() != "" {
		tmplMgr = tmplMgr.WithRepositorySrc(prj.GetRepository())
	}

	// Get template
	tpl, err := tmplMgr.Get(prj.GetTemplate())
	if err != nil {
		return err
	}

	// Sync
	snc.SetFileHook(snc.TemplateHook(prj.GetOptions()))

	for _, u := range tpl.GetSync() {
		srcFs := tpl.GetFs()
		if u.Template != "" {
			srcTpl, err := tmplMgr.Get(u.Template)
			if err != nil {
				return err
			}
			srcFs = srcTpl.GetFs()
		}
		err = snc.Sync(u.Destination, prj.GetFs(), u.Source, srcFs)
		if err != nil {
			return err
		}
	}

	return nil
}
