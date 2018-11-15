package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
)

var config = viper.New()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "manala",
	Short: "Let your projects plumbings up to date",
	Long: `Manala synchronize some boring parts of your projects,
such as makefile targets, virtualization and provisioning files...

Templates are pulled from git repository.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {

	// Version is set by public property instead of cobra.Command constructor,
	// so that it can be injected by main package
	rootCmd.Version = version

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("cache-dir", "d", "", "cache dir (default $HOME/.manala/cache)")

	config.BindPFlag("cache-dir", rootCmd.PersistentFlags().Lookup("cache-dir"))
}

func initConfig() {
	// Cache dir
	if config.GetString("cache-dir") == "" {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		config.Set("cache-dir", path.Join(home, ".manala", "cache"))
	}
}
