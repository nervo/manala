package cmd

import (
	"fmt"
	"github.com/apex/log"
	"github.com/fgrosse/goldi"
	"github.com/spf13/cobra"
	"manala/pkg/config"
	"manala/pkg/repository"
	"manala/pkg/template"
)

/*********/
/* Cobra */
/*********/

func ListCobra(container *goldi.Container) *cobra.Command {

	var opt listOptions

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List templates",
		Long: `List (manala list) will list templates available on
repository.

Example: manala list -> resulting in a template list display`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			container.MustGet("cmd.list").(*list).run(opt)
		},
	}

	return cmd
}

/***********/
/* Command */
/***********/

type listOptions struct {
}

func NewList(repositoryStore repository.StoreInterface, config *config.Config, logger log.Interface) *list {
	return &list{
		repositoryStore: repositoryStore,
		config:          config,
		logger:          logger,
	}
}

type list struct {
	repositoryStore repository.StoreInterface
	config          *config.Config
	logger          log.Interface
}

func (cmd *list) run(opt listOptions) {
	// Get repository
	rep, err := cmd.repositoryStore.Get(cmd.config.Repository)
	if err != nil {
		cmd.logger.WithError(err).Fatal("Error getting repository")
	}

	cmd.logger.WithField("src", rep.GetSrc()).Info("Repository gotten")

	err = rep.Walk(func(tpl template.Interface) {
		fmt.Printf("%s: %s\n", tpl.GetName(), tpl.GetDescription())
	})

	if err != nil {
		cmd.logger.WithError(err).Fatal("Error walking into templates")
	}
}
