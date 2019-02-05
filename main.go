package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/fgrosse/goldi"
	"github.com/fgrosse/goldi/validation"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"manala/cmd"
	"manala/pkg/config"
	"manala/pkg/project"
	"manala/pkg/repository"
	"manala/pkg/syncer"
	"manala/pkg/template"
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
	// Log
	log.SetHandler(cli.Default)

	// Root command
	rootCmd := &cobra.Command{
		Use:   "manala",
		Short: "Let your projects plumbings up to date",
		Long: `Manala synchronize some boring parts of your projects,
such as makefile targets, virtualization and provisioning files...

Templates are pulled from git repository.`,
		Version: version,
	}

	rootCmd.PersistentFlags().StringP("repository", "p", cfg.Repository, "repository")
	rootCmd.PersistentFlags().StringP("cache-dir", "c", cfg.CacheDir, "cache dir (default \"$HOME/.manala/cache\")")
	rootCmd.PersistentFlags().BoolP("debug", "d", cfg.Debug, "debug")

	// Container
	container := goldi.NewContainer(goldi.NewTypeRegistry(), map[string]interface{}{})

	// Commands
	rootCmd.AddCommand(cmd.UpdateCobra(container))
	rootCmd.AddCommand(cmd.WatchCobra(container))
	rootCmd.AddCommand(cmd.ListCobra(container))
	rootCmd.AddCommand(cmd.InitCobra(container))

	// Initialize
	cobra.OnInitialize(func() {
		// Logger
		logger := &log.Logger{
			Handler: cli.Default,
			Level:   log.InfoLevel,
		}

		// Viper
		vpr := viper.New()
		vpr.SetEnvPrefix("manala")
		vpr.AutomaticEnv()
		_ = vpr.BindPFlag("repository", rootCmd.PersistentFlags().Lookup("repository"))
		_ = vpr.BindPFlag("cache_dir", rootCmd.PersistentFlags().Lookup("cache-dir"))
		_ = vpr.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

		// Config
		err := vpr.Unmarshal(&cfg)
		if err != nil {
			logger.WithError(err).Fatal("Error unmarshalling config")
		}

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

		logger.WithField("repository", cfg.Repository).Debug("Config")
		logger.WithField("cache_dir", cfg.CacheDir).Debug("Config")
		logger.WithField("debug", cfg.Debug).Debug("Config")

		// File System
		fs := afero.NewOsFs()

		// Container
		container.RegisterAll(map[string]goldi.TypeFactory{
			"logger":             goldi.NewInstanceType(logger),
			"fs":                 goldi.NewInstanceType(fs),
			"project.manager":    goldi.NewType(project.NewManager, "@fs", "@logger"),
			"repository.manager": goldi.NewType(repository.NewManager, "@fs", "@logger", path.Join(cfg.CacheDir, "repository"), cfg.Debug),
			"template.manager":   goldi.NewType(template.NewSingleRepositoryManager, "@repository.manager", "@logger", cfg.Repository),
			"syncer":             goldi.NewType(syncer.New, "@logger"),
			"cmd.update":         goldi.NewStructType(cmd.UpdateCmd{}, "@project.manager", "@template.manager", "@syncer", "@logger"),
			"cmd.watch":          goldi.NewStructType(cmd.WatchCmd{}, "@project.manager", "@template.manager", "@syncer", "@logger"),
			"cmd.list":           goldi.NewStructType(cmd.ListCmd{}, "@template.manager", "@logger"),
			"cmd.init":           goldi.NewStructType(cmd.InitCmd{}, "@project.manager", "@template.manager", "@syncer", "@logger"),
		})

		val := validation.NewContainerValidator()
		val.MustValidate(container)
	})

	// Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
