package cmd

import (
	"github.com/apex/log"
	"github.com/fgrosse/goldi"
	"github.com/mostafah/fsync"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp/sideband"
	. "manala/pkg/config"
	"manala/pkg/project"
	"os"
	"path"
	"path/filepath"
)

/*******************/
/* Command - Cobra */
/*******************/

func UpdateCommandCobra(container *goldi.Container, config *Config) *cobra.Command {

	var options UpdateCommandOptions

	command := &cobra.Command{
		Use:   "update",
		Short: "Update project",
		Long: `Update (manala update) will update project, based on
template and related options defined in manala.yaml.

A optional dir could be passed as argument.

Example: manala update -> resulting in an update in current directory
Example: manala update /foo/bar -> resulting in an update in /foo/bar directory`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dir := ""
			if len(args) > 0 {
				dir = args[0]
			}
			container.MustGet("command.update").(UpdateCommandInterface).Run(dir, options, config)
		},
	}

	command.Flags().BoolVarP(&options.Recursive, "recursive", "r", false, "Recursive")

	return command
}

/***********/
/* Command */
/***********/

type UpdateCommandOptions struct {
	Recursive bool
}

type UpdateCommandInterface interface {
	Run(dir string, options UpdateCommandOptions, config *Config)
}

type UpdateCommand struct {
	ProjectFinder project.FinderInterface
	Logger        log.Interface
}

func (command *UpdateCommand) Run(dir string, options UpdateCommandOptions, config *Config) {

	var err error

	if dir != "" {
		if !filepath.IsAbs(dir) {
			command.Logger.WithField("dir", dir).Debug("Get absolute directory")
			if dir, err = filepath.Abs(dir); err != nil {
				command.Logger.WithError(err).Fatal("Error getting absolute directory")
			}
		}
	} else {
		command.Logger.Debug("Get working directory")
		dir, err = os.Getwd()
		if dir, err = os.Getwd(); err != nil {
			command.Logger.WithError(err).Fatal("Error getting working directory")
		}
	}

	// Project
	var p *project.Project

	if p, err = command.ProjectFinder.Find(dir); err != nil {
		command.Logger.WithError(err).Fatal("Error finding project")
	}

	command.Logger.WithField("dir", p.Dir).WithField("template", p.GetTemplate()).Info("Project found")

	repositoryUrl := "git@github.com:nervo/manala-templates.git"

	repositoryDir := path.Join(config.CacheDir, "repository")

	log.WithField("dir", repositoryDir).Info("Open repository")

	// Send git progress human readable information to stdout if debug enabled
	gitProgress := sideband.Progress(nil)

	if config.Debug {
		gitProgress = os.Stdout
	}

	repository, err := git.PlainOpen(repositoryDir)

	if err != nil {
		switch err {
		case git.ErrRepositoryNotExists:
			log.WithFields(log.Fields{
				"url": repositoryUrl,
				"dir": repositoryDir,
			}).Info("Clone repository")

			repository, err = git.PlainClone(repositoryDir, false, &git.CloneOptions{
				URL:               repositoryUrl,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				Progress:          gitProgress,
			})

			if err != nil {
				log.WithError(err).Fatal("Error cloning repository")
			}
		default:
			log.WithError(err).Fatal("Error opening repository")
		}
	} else {
		repositoryWorktree, err := repository.Worktree()

		if err != nil {
			log.WithError(err).Fatal("Error getting repository worktree")
		}

		log.WithField("dir", repositoryDir).Info("Pull repository worktree")

		err = repositoryWorktree.Pull(&git.PullOptions{
			RemoteName: "origin",
			Progress:   gitProgress,
		})

		if err != nil {
			switch err {
			case git.NoErrAlreadyUpToDate:
			default:
				log.WithError(err).Fatal("Error pulling repository worktree")
			}
		}
	}

	syncer := fsync.NewSyncer()
	syncer.Delete = true

	log.Info("Sync project")

	err = syncer.Sync(filepath.Join(p.Dir, ".manala"), filepath.Join(repositoryDir, p.GetTemplate(), ".manala"))
	if err != nil {
		log.WithError(err).Fatal("Error syncing project")
	}
}
