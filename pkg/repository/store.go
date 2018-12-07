package repository

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/apex/log"
	"github.com/spf13/afero"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp/sideband"
	"manala/pkg/template"
	"os"
	"path"
)

type StoreInterface interface {
	Get(src string) (Interface, error)
}

func NewStore(fs afero.Fs, templateFactory template.FactoryInterface, logger log.Interface, cacheDir string, debug bool) StoreInterface {
	return &store{
		fs:              fs,
		templateFactory: templateFactory,
		logger:          logger,
		repositories:    make(map[string]Interface),
		cacheDir:        cacheDir,
		debug:           debug,
	}
}

type store struct {
	fs              afero.Fs
	templateFactory template.FactoryInterface
	logger          log.Interface
	repositories    map[string]Interface
	cacheDir        string
	debug           bool
}

func (str *store) Get(src string) (Interface, error) {
	str.logger.WithField("src", src).Debug("Getting repository...")

	// Check if repository already in store
	if r, ok := str.repositories[src]; ok {
		str.logger.WithField("src", src).Debug("Returning repository from store...")
		return r, nil
	}

	// Send git progress human readable information to stdout if debug enabled
	gitProgress := sideband.Progress(nil)
	if str.debug {
		gitProgress = os.Stdout
	}

	hash := md5.New()
	hash.Write([]byte(src))

	// Repository cache directory should be unique
	dir := path.Join(str.cacheDir, "repository", hex.EncodeToString(hash.Sum(nil)))

	str.logger.WithField("dir", dir).Debug("Opening cache repository...")

	gitRepository, err := git.PlainOpen(dir)

	if err != nil {
		switch err {
		case git.ErrRepositoryNotExists:
			str.logger.Debug("Cloning cache git repository...")

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
		str.logger.Debug("Getting cache git repository worktree...")

		gitRepositoryWorktree, err := gitRepository.Worktree()

		if err != nil {
			return nil, ErrInvalid
		}

		str.logger.Debug("Pulling cache git repository worktree...")

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

	// Instantiate repository
	rep := New(
		src,
		afero.NewBasePathFs(str.fs, dir),
		str.templateFactory,
		str.logger,
	)

	// Store repository
	str.repositories[src] = rep

	return rep, nil
}
