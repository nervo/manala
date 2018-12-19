package cmd

import (
	"github.com/apex/log"
	"github.com/fgrosse/goldi"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"manala/pkg/project"
	"manala/pkg/syncer"
	"manala/pkg/template"
	"path"
	"path/filepath"
)

/*********/
/* Cobra */
/*********/

func WatchCobra(container *goldi.Container) *cobra.Command {

	var opt watchOptions

	cmd := &cobra.Command{
		Use:     "watch [DIR]",
		Aliases: []string{"wt"},
		Short:   "Watch project",
		Long: `Watch (manala watch) will watch project, and launch update on changes.

A optional dir could be passed as argument.

Example: manala watch -> resulting in an watch in current directory
Example: manala watch /foo/bar -> resulting in an watch in /foo/bar directory`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				args = append(args, "")
			}
			container.MustGet("cmd.watch").(*watch).run(args[0], opt)
		},
	}

	return cmd
}

/***********/
/* Command */
/***********/

type watchOptions struct {
}

func NewWatch(projectManager project.ManagerInterface, templateManager template.ManagerInterface, syncer syncer.Interface, logger log.Interface) *watch {
	return &watch{
		projectManager:  projectManager,
		templateManager: templateManager,
		syncer:          syncer,
		logger:          logger,
	}
}

type watch struct {
	projectManager  project.ManagerInterface
	templateManager template.ManagerInterface
	syncer          syncer.Interface
	logger          log.Interface
}

func (cmd *watch) run(dir string, opt watchOptions) {
	// Get real directory
	dir, err := getRealDir(dir)
	if err != nil {
		cmd.logger.WithError(err).Fatal("Error getting real directory")
	}

	// Find project
	prj, err := cmd.projectManager.Find(dir)
	if err != nil {
		cmd.logger.WithError(err).Fatal("Error finding project")
	}

	cmd.logger.WithFields(log.Fields{
		"template":   prj.GetTemplate(),
		"repository": prj.GetRepository(),
	}).Info("Project found")

	// Get project directory
	var prjDir string

	switch prj.GetFs().(type) {
	case *afero.BasePathFs:
		prjDir, err = prj.GetFs().(*afero.BasePathFs).RealPath("")
		if err != nil {
			cmd.logger.WithError(err).Fatal("Error getting project directory")
		}
	default:
		cmd.logger.Fatal("Project file system watching not supported")
	}

	// Get project supported config files
	cfgFiles := project.GetSupportedConfigFiles()
	for i, file := range cfgFiles {
		cfgFiles[i] = path.Join(prjDir, file)
	}

	// Watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		cmd.logger.WithError(err).Fatal("Error creating watcher")
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					for _, file := range cfgFiles {
						if filepath.Clean(event.Name) == file {
							cmd.logger.WithField("file", file).Info("Project config modified")
							prj, err = cmd.projectManager.Create(prj.GetFs())
							if err != nil {
								cmd.logger.WithError(err).Error("Error creating project")
							} else {
								err = cmd.projectManager.Sync(prj, cmd.templateManager, cmd.syncer)
								if err != nil {
									cmd.logger.WithError(err).Error("Error syncing project")
								} else {
									cmd.logger.Info("Project synced")
								}
							}
							break
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				cmd.logger.WithError(err).Error("Watching error")
			}
		}
	}()

	err = watcher.Add(prjDir)
	if err != nil {
		cmd.logger.WithError(err).Fatal("Error watching")
	}

	cmd.logger.Info("Start watching...")
	<-done
}
