package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/fgrosse/goldi"
	"github.com/fgrosse/goldi/validation"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"manala/cmd"
	"manala/pkg/config"
	"manala/pkg/project"
	"manala/pkg/sync"
	"manala/pkg/repository"
	"os"
	"path"
)

// Set at build time, by goreleaser, via ldflags
var version = "dev"

// Config
var cfg = &config.Config{
	Debug:      false,
	CacheDir:   "",
	Repository: "git@github.com:nervo/manala-templates.git",
}

func main() {
	// Logger
	logger := &log.Logger{
		Handler: cli.Default,
		Level:   log.InfoLevel,
	}

	// File System
	fs := afero.NewOsFs()

	// Container
	container := goldi.NewContainer(goldi.NewTypeRegistry(), map[string]interface{}{})
	container.RegisterAll(map[string]goldi.TypeFactory{
		"config":           goldi.NewInstanceType(cfg),
		"logger":           goldi.NewInstanceType(logger),
		"fs":               goldi.NewInstanceType(fs),
		"project.factory":  goldi.NewType(project.NewFactory, "@fs", "@logger"),
		"project.finder":   goldi.NewType(project.NewFinder, "@fs", "@project.factory", "@logger"),
		"repository.store": goldi.NewType(repository.NewStore, "@config", "@fs", "@logger"),
		"sync":             goldi.NewType(sync.NewSync),
		"cmd.update":       goldi.NewType(cmd.NewUpdate, "@project.finder", "@repository.store", "@sync", "@config", "@logger"),
	})

	val := validation.NewContainerValidator()
	val.MustValidate(container)

	// Root rootCmd
	rootCmd := &cobra.Command{
		Use:   "manala",
		Short: "Let your projects plumbings up to date",
		Long: `Manala synchronize some boring parts of your projects,
such as makefile targets, virtualization and provisioning files...

Templates are pulled from git repository.`,
		Version: version,
	}
	rootCmd.PersistentFlags().StringVarP(&cfg.Repository, "repository", "t", cfg.Repository, "repository")
	rootCmd.PersistentFlags().StringVarP(&cfg.CacheDir, "cache-dir", "c", cfg.CacheDir, "cache dir (default \"$HOME/.manala/cache\")")
	rootCmd.PersistentFlags().BoolVarP(&cfg.Debug, "debug", "d", cfg.Debug, "debug")

	// Initialize
	cobra.OnInitialize(func() {
		// Debug
		if cfg.Debug {
			logger.Level = log.DebugLevel
		}
		// Cache dir
		if cfg.CacheDir == "" {
			home, err := homedir.Dir()
			if err != nil {
				logger.WithError(err).Fatal("Error getting homedir")
			}
			cfg.CacheDir = path.Join(home, ".manala", "cache")
		}
	})

	// Command - Update
	rootCmd.AddCommand(cmd.UpdateCobra(container))

	// Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
