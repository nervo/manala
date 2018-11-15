package cmd

import (
	"fmt"
	"github.com/mostafah/fsync"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/src-d/go-git.v4"
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
		} else {
			dir, err = os.Getwd()
			if dir, err = os.Getwd(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		fmt.Printf("Dir: %s\n", dir)

		// Project config
		projectConfig := viper.New()
		projectConfig.SetConfigName("manala")
		projectConfig.SetConfigType("yaml")
		projectConfig.AddConfigPath(dir)

		if err = projectConfig.ReadInConfig(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		template := projectConfig.GetString("manala.template")

		if template == "" {
			fmt.Println("Manala template is not defined or empty")
			os.Exit(1)
		}

		fmt.Printf("Template: %s\n", template)

		repositoryUrl := "git@github.com:nervo/manala-templates.git"
		repositoryDir := path.Join(config.GetString("cache-dir"), "repository")

		fmt.Printf("Repository dir: %s\n", repositoryDir)

		repository, err := git.PlainOpen(repositoryDir)

		if err != nil {
			switch err {
			case git.ErrRepositoryNotExists:
				fmt.Printf("Clone \"%s\" into \"%s\"\n", repositoryUrl, repositoryDir)

				repository, err = git.PlainClone(repositoryDir, false, &git.CloneOptions{
					URL:               repositoryUrl,
					RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
					Progress:          os.Stdout,
				})

				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			default:
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			repositoryWorktree, err := repository.Worktree()

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = repositoryWorktree.Pull(&git.PullOptions{RemoteName: "origin"})

			if err != nil {
				switch err {
				case git.NoErrAlreadyUpToDate:
				default:
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}

		syncer := fsync.NewSyncer()
		syncer.Delete = true

		err = syncer.Sync(filepath.Join(dir, ".manala"), filepath.Join(repositoryDir, template, ".manala"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
