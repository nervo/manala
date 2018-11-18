package cmd

import (
	"github.com/apex/log"
	"github.com/mostafah/fsync"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp/sideband"
	"os"
	"path"
	"path/filepath"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update project",
	Long: `Update (manala update) will update project, based on
template and related options defined in manala.yaml.

A optional dir could be passed as argument.

Example: manala update -> resulting in an update in current directory
Example: manala update /foo/bar -> resulting in an update in /foo/bar directory`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		var err error

		// Dir is either specified by first argument
		// or working directory
		var dir string

		if len(args) > 0 {
			dir = args[0]
			if !filepath.IsAbs(dir) {
				log.WithField("dir", dir).Debug("Get absolute directory")
				if dir, err = filepath.Abs(dir); err != nil {
					log.WithError(err).Fatal("Error getting absolute directory")
				}
			}
		} else {
			log.Debug("Get working directory")
			dir, err = os.Getwd()
			if dir, err = os.Getwd(); err != nil {
				log.WithError(err).Fatal("Error getting working directory")
			}
		}

		log.WithField("dir", dir).Info("Set directory")

		// Project config
		projectConfig := viper.New()
		projectConfig.SetConfigName("manala")
		projectConfig.SetConfigType("yaml")
		projectConfig.AddConfigPath(dir)

		if err = projectConfig.ReadInConfig(); err != nil {
			log.WithError(err).Fatal("Error reading project configuration")
		}

		template := projectConfig.GetString("manala.template")

		if template == "" {
			log.Fatal("Manala template is not defined or empty")
		}

		log.WithField("template", template).Info("Set template")

		repositoryUrl := "git@github.com:nervo/manala-templates.git"
		repositoryDir := path.Join(config.GetString("cache-dir"), "repository")

		log.WithField("dir", repositoryDir).Info("Open repository")

		// Send git progress human readable information to stdout if debug enabled
		gitProgress := sideband.Progress(nil)
		if config.GetBool("debug") {
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

		err = syncer.Sync(filepath.Join(dir, ".manala"), filepath.Join(repositoryDir, template, ".manala"))
		if err != nil {
			log.WithError(err).Fatal("Error syncing project")
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
