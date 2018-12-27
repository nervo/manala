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

	var opt UpdateOptions

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
			container.MustGet("cmd.update").(*UpdateCmd).Run(args[0], opt)
		},
	}

	cmd.Flags().BoolVarP(&opt.Recursive, "recursive", "r", false, "Recursive")

	return cmd
}

/***********/
/* Options */
/***********/

type UpdateOptions struct {
	Recursive bool
}

/***********/
/* Command */
/***********/

type UpdateCmd struct {
	ProjectManager  project.ManagerInterface
	TemplateManager template.ManagerInterface
	Syncer          syncer.Interface
	Logger          log.Interface
}

func (cmd *UpdateCmd) Run(dir string, opt UpdateOptions) {
	// Get real directory
	dir, err := getRealDir(dir)
	if err != nil {
		cmd.Logger.WithError(err).Fatal("Error getting real directory")
	}

	if opt.Recursive {
		// Recursively find projects
		err = cmd.ProjectManager.Walk(dir, func(prj *project.ManagedProject) {
			cmd.Logger.WithFields(log.Fields{
				"template":   prj.GetTemplate(),
				"repository": prj.GetRepository(),
			}).Info("Project found")

			// Sync
			err = cmd.syncProject(prj)
			if err != nil {
				cmd.Logger.WithError(err).Fatal("Error syncing project")
			}
		})
		if err != nil {
			cmd.Logger.WithError(err).Fatal("Error finding projects recursively")
		}
	} else {
		// Find project
		prj, err := cmd.ProjectManager.Find(dir)
		if err != nil {
			cmd.Logger.WithError(err).Fatal("Error finding project")
		}

		cmd.Logger.WithFields(log.Fields{
			"template":   prj.GetTemplate(),
			"repository": prj.GetRepository(),
		}).Info("Project found")

		// Sync
		err = cmd.syncProject(prj)
		if err != nil {
			cmd.Logger.WithError(err).Fatal("Error syncing project")
		}
	}
}

func (cmd *UpdateCmd) syncProject(prj project.Interface) error {
	tmplMgr := cmd.TemplateManager

	// Custom project repository
	if prj.GetRepository() != "" {
		tmplMgr = tmplMgr.WithRepositorySrc(prj.GetRepository())
	}

	err := cmd.Syncer.SyncProject(prj, tmplMgr)
	if err != nil {
		return err
	}

	cmd.Logger.Info("Project synced")

	return nil
}
