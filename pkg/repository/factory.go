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

type FactoryInterface interface {
	Create(src string) (Interface, error)
}

func NewFactory(fs afero.Fs, templateFactory template.FactoryInterface, logger log.Interface, cacheDir string, debug bool) FactoryInterface {
	return &factory{
		fs:              fs,
		templateFactory: templateFactory,
		logger:          logger,
		cacheDir:        cacheDir,
		debug:           debug,
	}
}

type factory struct {
	fs              afero.Fs
	templateFactory template.FactoryInterface
	logger          log.Interface
	cacheDir        string
	debug           bool
}

func (fa *factory) Create(src string) (Interface, error) {
	// Send git progress human readable information to stdout if debug enabled
	gitProgress := sideband.Progress(nil)
	if fa.debug {
		gitProgress = os.Stdout
	}

	hash := md5.New()
	hash.Write([]byte(src))

	// Repository cache directory should be unique
	dir := path.Join(fa.cacheDir, hex.EncodeToString(hash.Sum(nil)))

	fa.logger.WithField("dir", fa).Debug("Opening cache repository...")

	gitRepository, err := git.PlainOpen(dir)

	if err != nil {
		switch err {
		case git.ErrRepositoryNotExists:
			fa.logger.Debug("Cloning cache git repository...")

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
		fa.logger.Debug("Getting cache git repository worktree...")

		gitRepositoryWorktree, err := gitRepository.Worktree()

		if err != nil {
			return nil, ErrInvalid
		}

		fa.logger.Debug("Pulling cache git repository worktree...")

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
		afero.NewBasePathFs(fa.fs, dir),
		fa.templateFactory,
		fa.logger,
	)

	return rep, nil
}
