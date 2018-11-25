package template

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/apex/log"
	"github.com/spf13/afero"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp/sideband"
	"manala/pkg/config"
	"os"
	"path"
)

var (
	ErrRepositoryUnclonable = errors.New("repository unclonable")
	ErrRepositoryUnopenable = errors.New("repository unopenable")
	ErrRepositoryInvalid    = errors.New("repository invalid")
)

type RepositoryStoreInterface interface {
	Get(src string) (RepositoryInterface, error)
}

func NewRepositoryStore(config *config.Config, fs afero.Fs, logger log.Interface) RepositoryStoreInterface {
	return &repositoryStore{
		config:       config,
		fs:           fs,
		logger:       logger,
		repositories: make(map[string]RepositoryInterface),
	}
}

type repositoryStore struct {
	config       *config.Config
	fs           afero.Fs
	logger       log.Interface
	repositories map[string]RepositoryInterface
}

func (s *repositoryStore) Get(src string) (RepositoryInterface, error) {
	s.logger.WithField("src", src).Debug("Getting template repository...")

	// Check if repository already in store
	if r, ok := s.repositories[src]; ok {
		s.logger.WithField("src", src).Debug("Returning template repository from store...")
		return r, nil
	}

	// Send git progress human readable information to stdout if debug enabled
	gitProgress := sideband.Progress(nil)
	if s.config.Debug {
		gitProgress = os.Stdout
	}

	hash := md5.New()
	hash.Write([]byte(src))

	// Template repository cache directory should be unique
	dir := path.Join(s.config.CacheDir, "template", "repository", hex.EncodeToString(hash.Sum(nil)))

	s.logger.WithField("dir", dir).Debug("Opening cache template repository...")

	gitRepository, err := git.PlainOpen(dir)

	if err != nil {
		switch err {
		case git.ErrRepositoryNotExists:
			log.Debug("Cloning cache template repository...")

			gitRepository, err = git.PlainClone(dir, false, &git.CloneOptions{
				URL:               src,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				Progress:          gitProgress,
			})

			if err != nil {
				return nil, ErrRepositoryUnclonable
			}
		default:
			return nil, ErrRepositoryUnopenable
		}
	} else {
		log.Debug("Getting repository worktree...")

		repositoryWorktree, err := gitRepository.Worktree()

		if err != nil {
			return nil, ErrRepositoryInvalid
		}

		log.Debug("Pulling repository worktree...")

		err = repositoryWorktree.Pull(&git.PullOptions{
			RemoteName: "origin",
			Progress:   gitProgress,
		})

		if err != nil {
			switch err {
			case git.NoErrAlreadyUpToDate:
			default:
				return nil, ErrRepositoryInvalid
			}
		}
	}

	// Instantiate repository
	r := &repository{
		dir: dir,
		fs:  s.fs,
	}

	// Store repository
	s.repositories[src] = r

	return r, nil
}
