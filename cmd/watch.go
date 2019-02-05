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
	"strings"
)

/*********/
/* Cobra */
/*********/

func WatchCobra(container *goldi.Container) *cobra.Command {

	var opt WatchOptions

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
			container.MustGet("cmd.watch").(*WatchCmd).Run(args[0], opt)
		},
	}

	cmd.Flags().BoolVarP(&opt.Template, "template", "t", false, "Template")
	cmd.Flags().BoolVarP(&opt.Notify, "notify", "n", false, "Notify")

	return cmd
}

/***********/
/* Options */
/***********/

type WatchOptions struct {
	Template bool
	Notify   bool
}

/***********/
/* Command */
/***********/

type WatchCmd struct {
	ProjectManager  project.ManagerInterface
	TemplateManager template.ManagerInterface
	Syncer          syncer.Interface
	Logger          log.Interface
}

func (cmd *WatchCmd) Run(dir string, opt WatchOptions) {
	// Get real directory
	dir, err := getRealDir(dir)
	if err != nil {
		cmd.Logger.WithError(err).Fatal("Error getting real directory")
	}

	// Find project
	prj, err := cmd.ProjectManager.Find(dir)
	if err != nil {
		cmd.Logger.WithError(err).Fatal("Error finding project")
	}

	cmd.Logger.WithFields(log.Fields{
		"template":   prj.GetTemplate(),
		"repository": prj.GetRepository(),
	}).Info("Project found")

	// Get project dir
	prjDir := prj.GetDir()

	// Get project config files
	var prjConfigFiles []string
	for _, file := range prj.GetSupportedConfigFiles() {
		prjConfigFiles = append(prjConfigFiles, file)
	}

	// Watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		cmd.Logger.WithError(err).Fatal("Error creating watcher")
	}
	defer watcher.Close()

	// Watch project
	err = watcher.Add(prjDir)
	if err != nil {
		cmd.Logger.WithError(err).Fatal("Error adding project watching")
	}

	// Get sync method function
	syncProject := cmd.syncProjectFunc(prjDir, watcher, opt.Template)

	err = syncProject()
	if err != nil {
		cmd.Logger.WithError(err).Fatal("Error syncing project")
	}

	cmd.Logger.Info("Project synced")

	cmd.Logger.Info("Start watching...")

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				cmd.Logger.WithField("event", event).Debug("Watch event")

				if event.Op != fsnotify.Chmod {
					modified := false
					dir := filepath.Dir(filepath.Clean(event.Name))
					// Modified directory is not project one. That could only means template's one
					if dir != prjDir {
						cmd.Logger.WithField("dir", dir).Info("Project template modified")
						modified = true
					} else {
						for _, file := range prjConfigFiles {
							if filepath.Clean(event.Name) == file {
								cmd.Logger.WithField("file", file).Info("Project config modified")
								modified = true
								break
							}
						}
					}

					if modified {
						err = syncProject()
						if err != nil {
							cmd.Logger.WithError(err).Error("Error syncing project")
							if opt.Notify {
								err = beeep.Alert("Manala", strings.Replace(err.Error(), `"`, `\"`, -1), "")
								if err != nil {
									cmd.Logger.WithError(err).Warn("Error notifying")
								}
							}
						} else {
							cmd.Logger.Info("Project synced")
							if opt.Notify {
								err := beeep.Notify("Manala", "Project synced", "")
								if err != nil {
									cmd.Logger.WithError(err).Warn("Error notifying")
								}
							}
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				cmd.Logger.WithError(err).Error("Watching error")
			}
		}
	}()
	<-done
}

func (cmd *WatchCmd) syncProjectFunc(dir string, watcher *fsnotify.Watcher, watchTemplate bool) func() error {
	var baseTmplDirs []string

	return func() error {
		// Create project from file system
		prj, err := cmd.ProjectManager.Get(dir)
		if err != nil {
			return err
		}

		tmplMgr := cmd.TemplateManager

		// Custom project repository
		if prj.GetRepository() != "" {
			tmplMgr = tmplMgr.WithRepositorySrc(prj.GetRepository())
		}

		if watchTemplate {
			// Get project template
			tmpl, err := tmplMgr.Get(prj.GetTemplate())
			if err != nil {
				return err
			} else {
				var tmplDirs []string
				err = filepath.Walk(tmpl.GetDir(), func(path string, info os.FileInfo, err error) error {
					if info.Mode().IsDir() {
						tmplDirs = append(tmplDirs, path)
						err = watcher.Add(path)
						if err != nil {
							return err
						}
					}

					return nil
				})
				if err != nil {
					return err
				}

				// Remove unneeded dirs from watching
				for _, baseDir := range baseTmplDirs {
					found := false
					for _, dir := range tmplDirs {
						if dir == baseDir {
							found = true
						}
					}
					if !found {
						watcher.Remove(baseDir)
					}
				}

				baseTmplDirs = tmplDirs
			}
		}

		err = cmd.Syncer.SyncProject(prj, tmplMgr)
		if err != nil {
			return err
		}

		return nil
	}
}
