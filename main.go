package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/fgrosse/goldi"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"manala/cmd"
	. "manala/pkg/config"
	"manala/pkg/project"
	"os"
	"path"
)

var (
	version = "dev" // Set at build time, by goreleaser, via ldflags
	config  = &Config{
		Debug:    false,
		CacheDir: "",
	}
)

func main() {
	// Logger
	logger := &log.Logger{
		Handler: cli.Default,
		Level:   log.InfoLevel,
	}

	// Container
	container := goldi.NewContainer(goldi.NewTypeRegistry(), map[string]interface{}{})
	container.RegisterAll(map[string]goldi.TypeFactory{
		"logger":          goldi.NewInstanceType(logger),
		"project.factory": goldi.NewStructType(new(project.Factory), "@logger"),
		"project.finder":  goldi.NewStructType(new(project.Finder), "@project.factory", "@logger"),
		"command.update":  goldi.NewStructType(new(cmd.UpdateCommand), "@project.finder", "@logger"),
	})

	// Command
	command := &cobra.Command{
		Use:     "manala",
		Version: version,
		Short:   "Let your projects plumbings up to date",
		Long: `Manala synchronize some boring parts of your projects,
such as makefile targets, virtualization and provisioning files...

Templates are pulled from git repository.`,
	}
	command.PersistentFlags().StringVarP(&config.CacheDir, "cache-dir", "c", "", "cache dir (default $HOME/.manala/cache)")
	command.PersistentFlags().BoolVarP(&config.Debug, "debug", "d", false, "debug")

	// Initialize
	cobra.OnInitialize(func() {
		// Debug
		if config.Debug {
			logger.Level = log.DebugLevel
		}
		// Cache dir
		if config.CacheDir == "" {
			home, err := homedir.Dir()
			if err != nil {
				logger.WithError(err).Fatal("Error getting homedir")
			}
			config.CacheDir = path.Join(home, ".manala", "cache")
		}
	})

	// Command - Update
	command.AddCommand(cmd.UpdateCommandCobra(container, config))

	// Execute
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
