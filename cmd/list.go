package cmd

import (
	"fmt"
	"github.com/apex/log"
	"github.com/fgrosse/goldi"
	"github.com/spf13/cobra"
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

func NewList(templateManager template.ManagerInterface, logger log.Interface) *list {
	return &list{
		templateManager: templateManager,
		logger:          logger,
	}
}

type list struct {
	templateManager template.ManagerInterface
	logger          log.Interface
}

func (cmd *list) run(opt listOptions) {
	// Walk
	err := cmd.templateManager.Walk(func(tpl template.Interface) {
		fmt.Printf("%s: %s\n", tpl.GetName(), tpl.GetDescription())
	})

	if err != nil {
		cmd.logger.WithError(err).Fatal("Error walking templates")
	}
}
