package repository

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/apex/log"
	"github.com/spf13/afero"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp/sideband"
	"os"
	"path"
	"path/filepath"
)

/**********************/
/* Managed Repository */
/**********************/

type ManagedRepository struct {
	Interface
	dir string
}

func (rep *ManagedRepository) GetDir() string {
	return rep.dir
}

/***********/
/* Manager */
/***********/

type ManagerInterface interface {
	Create(src string) (*ManagedRepository, error)
}

func NewManager(fs afero.Fs, logger log.Interface, cacheDir string, debug bool) *manager {
	return &manager{
		fs:       fs,
		logger:   logger,
		cacheDir: cacheDir,
		debug:    debug,
	}
}

type manager struct {
	fs       afero.Fs
	logger   log.Interface
	cacheDir string
	debug    bool
}

func (mgr *manager) Create(src string) (*ManagedRepository, error) {
	switch {
	case filepath.Ext(src) == ".git":
		return mgr.createGit(src)
	}

	return mgr.createDirectory(src)
}

func (mgr *manager) createDirectory(src string) (*ManagedRepository, error) {
	// Todo: ensure src exists...

	// Instantiate repository
	return &ManagedRepository{
		Interface: &repository{
			src: src,
			fs:  afero.NewBasePathFs(mgr.fs, src),
		},
		dir: src,
	}, nil
}

func (mgr *manager) createGit(src string) (*ManagedRepository, error) {
	// Send git progress human readable information to stdout if debug enabled
	gitProgress := sideband.Progress(nil)
	if mgr.debug {
		gitProgress = os.Stdout
	}

	hash := md5.New()
	hash.Write([]byte(src))

	// Repository cache directory should be unique
	dir := path.Join(mgr.cacheDir, hex.EncodeToString(hash.Sum(nil)))

	mgr.logger.WithField("dir", mgr).Debug("Opening cache repository...")

	gitRepository, err := git.PlainOpen(dir)

	if err != nil {
		switch err {
		case git.ErrRepositoryNotExists:
			mgr.logger.Debug("Cloning cache git repository...")

			gitRepository, err = git.PlainClone(dir, false, &git.CloneOptions{
				URL:               src,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				Progress:          gitProgress,
			})

			if err != nil {
				return nil, ErrUnclonable
			}
		default:
			return nil, ErrUnopenable
		}
	} else {
		mgr.logger.Debug("Getting cache git repository worktree...")

		gitRepositoryWorktree, err := gitRepository.Worktree()

		if err != nil {
			return nil, ErrInvalid
		}

		mgr.logger.Debug("Pulling cache git repository worktree...")

		err = gitRepositoryWorktree.Pull(&git.PullOptions{
			RemoteName: "origin",
			Progress:   gitProgress,
		})

		if err != nil {
			switch err {
			case git.NoErrAlreadyUpToDate:
			default:
				return nil, err
			}
		}
	}

	return &ManagedRepository{
		Interface: &repository{
			src: src,
			fs:  afero.NewBasePathFs(mgr.fs, dir),
		},
		dir: dir,
	}, nil
}
