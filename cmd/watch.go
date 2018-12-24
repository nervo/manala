package cmd

import (
	"github.com/apex/log"
	"github.com/fgrosse/goldi"
	"github.com/fsnotify/fsnotify"
	"github.com/gen2brain/beeep"
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

	cmd.Flags().BoolVarP(&opt.Template, "template", "t", false, "Template")
	cmd.Flags().BoolVarP(&opt.Notify, "notify", "n", false, "Notify")

	return cmd
}

/***********/
/* Command */
/***********/

type watchOptions struct {
	Template bool
	Notify   bool
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

	var prjConfigFiles []string

	for _, file := range prj.GetSupportedConfigFiles() {
		prjConfigFiles = append(prjConfigFiles, file)
	}

	tmplMgr := cmd.templateManager

	// Custom project repository
	if prj.GetRepository() != "" {
		tmplMgr = cmd.templateManager.WithRepositorySrc(prj.GetRepository())
	}

	var tmplDirs []string

	// Get template directories
	if opt.Template {
		// Get project template
		tmpl, err := tmplMgr.Get(prj.GetTemplate())
		if err != nil {
			cmd.logger.WithError(err).Warn("Error getting project template")
		} else {
			err = filepath.Walk(tmpl.GetDir(), func(path string, info os.FileInfo, err error) error {
				if info.Mode().IsDir() {
					tmplDirs = append(tmplDirs, path)
				}

				return nil
			})
			if err != nil {
				cmd.logger.WithError(err).Warn("Error walking into project template")
			}
		}
	}

	// Watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		cmd.logger.WithError(err).Fatal("Error creating watcher")
	}
	defer watcher.Close()

	// Watch project
	err = watcher.Add(prj.GetDir())
	if err != nil {
		cmd.logger.WithError(err).Fatal("Error adding project watching")
	}

	// Watch template
	for _, dir := range tmplDirs {
		err = watcher.Add(dir)
		if err != nil {
			cmd.logger.WithError(err).Fatal("Error adding templaten watching")
		}
	}

	cmd.logger.Info("Start watching...")

	done := make(chan bool)
	go func() {
		prj, err := cmd.projectManager.Create(prj.GetFs())
		if err != nil {
			cmd.logger.WithError(err).Fatal("Error creating project")
		}
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					modified := false
					for _, file := range prjConfigFiles {
						if filepath.Clean(event.Name) == file {
							cmd.logger.WithField("file", file).Info("Project config modified")

							modifiedPrj, err := cmd.projectManager.Create(prj.GetFs())
							if err != nil {
								cmd.logger.WithError(err).Error("Error creating project")
							} else {
								modified = true
								prj = modifiedPrj
							}
							break
						}
					}
					for _, dir := range tmplDirs {
						if filepath.Dir(filepath.Clean(event.Name)) == dir {
							cmd.logger.WithField("dir", dir).Info("Project template modified")
							modified = true
						}
					}

					if modified {
						err := cmd.syncer.SyncProject(prj, tmplMgr)
						if err != nil {
							cmd.logger.WithError(err).Error("Error syncing project")
						} else {
							cmd.logger.Info("Project synced")
							if opt.Notify {
								_ = beeep.Notify("Manala", "Project synced", "")
							}
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
	<-done
}
