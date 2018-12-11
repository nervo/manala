package cmd

import (
	"github.com/apex/log"
	"github.com/fgrosse/goldi"
	"github.com/spf13/cobra"
	"manala/pkg/project"
	"manala/pkg/syncer"
	"manala/pkg/template"
	"os"
	"path/filepath"
)

/*********/
/* Cobra */
/*********/

func UpdateCobra(container *goldi.Container) *cobra.Command {

	var opt updateOptions

	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"up"},
		Short:   "Update project",
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
			container.MustGet("cmd.update").(*update).run(dir, opt)
		},
	}

	cmd.Flags().BoolVarP(&opt.Recursive, "recursive", "r", false, "Recursive")

	return cmd
}

/***********/
/* Command */
/***********/

type updateOptions struct {
	Recursive bool
}

func NewUpdate(projectManager project.ManagerInterface, templateManager template.ManagerInterface, syncer syncer.Interface, logger log.Interface) *update {
	return &update{
		projectManager:  projectManager,
		templateManager: templateManager,
		syncer:          syncer,
		logger:          logger,
	}
}

type update struct {
	projectManager  project.ManagerInterface
	templateManager template.ManagerInterface
	syncer          syncer.Interface
	logger          log.Interface
}

func (cmd *update) run(dir string, opt updateOptions) {

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

	if opt.Recursive {
		err = cmd.projectManager.Walk(dir, cmd.updateProject)
		if err != nil {
			cmd.logger.WithError(err).Fatal("Error finding projects recursively")
		}
	} else {
		prj, err := cmd.projectManager.Find(dir)
		if err != nil {
			cmd.logger.WithError(err).Fatal("Error finding project")
		}

		cmd.updateProject(prj)
	}
}

func (cmd *update) updateProject(prj project.Interface) {
	cmd.logger.WithField("template", prj.GetTemplate()).Info("Project found")

	templateManager := cmd.templateManager

	// Custom project repository
	if prj.GetRepository() != "" {
		templateManager = templateManager.WithRepositorySrc(prj.GetRepository())
	}

	// Get template
	tpl, err := templateManager.Get(prj.GetTemplate())
	if err != nil {
		cmd.logger.WithError(err).Fatal("Error getting template")
	}

	// Sync
	for _, path := range tpl.GetSync() {
		err = cmd.syncer.Sync(path, prj.GetFs(), tpl.GetFs())
		if err != nil {
			cmd.logger.WithError(err).Fatal("Error syncing project")
		}
	}

	cmd.logger.Info("Project synced")
}
