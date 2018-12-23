package cmd

import (
	"github.com/apex/log"
	"github.com/fgrosse/goldi"
	"github.com/spf13/cobra"
	"manala/pkg/project"
	"manala/pkg/syncer"
	"manala/pkg/template"
)

/*********/
/* Cobra */
/*********/

func UpdateCobra(container *goldi.Container) *cobra.Command {

	var opt updateOptions

	cmd := &cobra.Command{
		Use:     "update [DIR]",
		Aliases: []string{"up"},
		Short:   "Update project",
		Long: `Update (manala update) will update project, based on
template and related options defined in manala.yaml.

A optional dir could be passed as argument.

Example: manala update -> resulting in an update in current directory
Example: manala update /foo/bar -> resulting in an update in /foo/bar directory`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				args = append(args, "")
			}
			container.MustGet("cmd.update").(*update).run(args[0], opt)
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
	// Get real directory
	dir, err := getRealDir(dir)
	if err != nil {
		cmd.logger.WithError(err).Fatal("Error getting real directory")
	}

	if opt.Recursive {
		err = cmd.projectManager.Walk(dir, func(prj *project.ManagedProject) {
			err = cmd.syncProject(prj, cmd.templateManager)
			if err != nil {
				cmd.logger.WithError(err).Fatal("Error syncing project")
			}
		})
		if err != nil {
			cmd.logger.WithError(err).Fatal("Error finding projects recursively")
		}
	} else {
		prj, err := cmd.projectManager.Find(dir)
		if err != nil {
			cmd.logger.WithError(err).Fatal("Error finding project")
		}

		err = cmd.syncProject(prj, cmd.templateManager)
		if err != nil {
			cmd.logger.WithError(err).Fatal("Error syncing project")
		}
	}
}

func (cmd *update) syncProject(prj project.Interface, tmplMgr template.ManagerInterface) error {
	cmd.logger.WithFields(log.Fields{
		"template":   prj.GetTemplate(),
		"repository": prj.GetRepository(),
	}).Info("Project found")

	// Custom project repository
	if prj.GetRepository() != "" {
		tmplMgr = cmd.templateManager.WithRepositorySrc(prj.GetRepository())
	}

	err := cmd.syncer.SyncProject(prj, tmplMgr)
	if err != nil {
		return err
	}

	cmd.logger.Info("Project synced")

	return nil
}
