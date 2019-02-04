package project

import (
	"errors"
	"github.com/apex/log"
	"github.com/asaskevich/govalidator"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"os"
	"path"
	"path/filepath"
)

/**********/
/* Errors */
/**********/

var (
	ErrNotFound = errors.New("project not found")
	ErrConfig   = errors.New("project config invalid")
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

/*******************/
/* Managed Project */
/*******************/

type ManagedProject struct {
	Interface
	dir string
}

func (prj *ManagedProject) GetDir() string {
	return prj.dir
}

func (prj *ManagedProject) GetSupportedConfigFiles() []string {
	var files []string
	for _, cfg := range supportedConfigNames {
		for _, ext := range viper.SupportedExts {
			files = append(files, path.Join(prj.GetDir(), cfg.name+"."+ext))
		}
	}

	return files
}

/***********/
/* Manager */
/***********/

type ManagerInterface interface {
	Create(fs afero.Fs) (*project, error)
	Get(dir string) (*ManagedProject, error)
	Find(dir string) (*ManagedProject, error)
	Walk(dir string, fn ManagerWalkFunc) error
}

func NewManager(fs afero.Fs, logger log.Interface) *manager {
	return &manager{
		fs:     fs,
		logger: logger,
	}
}

type manager struct {
	fs     afero.Fs
	logger log.Interface
}

func (mgr *manager) Create(fs afero.Fs) (*project, error) {
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

	var cfg Config

	// Unmarshalling
	if err := vpr.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Validation
	if _, err := govalidator.ValidateStruct(cfg); err != nil {
		return nil, err
	}

	return &project{
		fs:      fs,
		config:  cfg,
		options: prjOptions,
	}, nil
}

// Get a project by dir
func (mgr *manager) Get(dir string) (*ManagedProject, error) {
	prj, err := mgr.Create(afero.NewBasePathFs(mgr.fs, dir))
	if err == nil {
		return &ManagedProject{
			Interface: prj,
			dir:       dir,
		}, nil
	}

	return nil, ErrNotFound
}

// Find a project by browsing dir then its parents up to root
func (mgr *manager) Find(dir string) (*ManagedProject, error) {
	// Todo: oh god... this algorithm sucks... how the hell git do ?
	lastDir := ""
	for dir != lastDir {
		lastDir = dir
		mgr.logger.WithField("dir", dir).Debug("Searching project...")
		prj, err := mgr.Get(dir)
		if err == nil {
			return prj, nil
		}
		dir = filepath.Dir(dir)
	}

	return nil, ErrNotFound
}

type ManagerWalkFunc func(project *ManagedProject)

// Find projects recursively starting from dir
func (mgr *manager) Walk(dir string, fn ManagerWalkFunc) error {

	err := afero.Walk(mgr.fs, dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		mgr.logger.WithField("dir", path).Debug("Searching project...")
		prj, err := mgr.Get(path)
		if err != nil {
			return nil
		}

		fn(prj)

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
