package cmd

import (
	"github.com/apex/log"
	"github.com/fgrosse/goldi"
	"github.com/spf13/cobra"
	"manala/pkg/config"
	"manala/pkg/project"
	"manala/pkg/sync"
	"manala/pkg/template"
	"os"
	"path/filepath"
)

/*********/
/* Cobra */
/*********/

func UpdateCobra(container *goldi.Container) *cobra.Command {

	var o updateOptions

	cmd := &cobra.Command{
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
			container.MustGet("cmd.update").(*update).Run(dir, o)
		},
	}

	cmd.Flags().BoolVarP(&o.Recursive, "recursive", "r", false, "Recursive")

	return cmd
}

/***********/
/* Command */
/***********/

type updateOptions struct {
	Recursive bool
}

func NewUpdate(projectFinder project.FinderInterface, templateRepositoryStore template.RepositoryStoreInterface, sync sync.Interface, config *config.Config, logger log.Interface) *update {
	return &update{
		projectFinder:           projectFinder,
		templateRepositoryStore: templateRepositoryStore,
		sync:                    sync,
		config:                  config,
		logger:                  logger,
	}
}

type update struct {
	projectFinder           project.FinderInterface
	templateRepositoryStore template.RepositoryStoreInterface
	sync                    sync.Interface
	config                  *config.Config
	logger                  log.Interface
}

func (cmd *update) Run(dir string, o updateOptions) {

	var err error

	if dir != "" {
		if !filepath.IsAbs(dir) {
			cmd.logger.WithField("dir", dir).Debug("Get absolute directory")
			dir, err = filepath.Abs(dir)
			if err != nil {
				cmd.logger.WithError(err).Fatal("Error getting absolute directory")
			}
		}
	} else {
		cmd.logger.Debug("Get working directory")
		dir, err = os.Getwd()
		if err != nil {
			cmd.logger.WithError(err).Fatal("Error getting working directory")
		}
	}

	// Find project
	p, err := cmd.projectFinder.Find(dir)
	if err != nil {
		cmd.logger.WithError(err).Fatal("Error finding project")
	}

	cmd.logger.WithField("dir", p.GetDir()).WithField("template", p.GetTemplate()).Info("Project found")

	// Get template repository
	r, err := cmd.templateRepositoryStore.Get(cmd.config.TemplateRepository)
	if err != nil {
		cmd.logger.WithError(err).Fatal("Error getting template repository")
	}

	cmd.logger.WithField("dir", r.GetDir()).Info("Template repository gotten")

	// Get template
	t, err := r.Get(p.GetTemplate())
	if err != nil {
		cmd.logger.WithError(err).Fatal("Error getting template")
	}

	cmd.logger.WithField("dir", t.GetDir()).Info("Template gotten")

	// Sync project
	err = cmd.sync.Sync(p, t)
	if err != nil {
		cmd.logger.WithError(err).Fatal("Error syncing project")
	}

	cmd.logger.Info("Project synced")
}
