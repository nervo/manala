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

	var opt ListOptions

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List templates",
		Long: `List (manala list) will list templates available on
repository.

Example: manala list -> resulting in a template list display`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			container.MustGet("cmd.list").(*ListCmd).Run(opt)
		},
	}

	return cmd
}

/***********/
/* Options */
/***********/

type ListOptions struct {
}

/***********/
/* Command */
/***********/

type ListCmd struct {
	TemplateManager template.ManagerInterface
	Logger          log.Interface
}

func (cmd *ListCmd) Run(opt ListOptions) {
	// Walk into templates
	err := cmd.TemplateManager.Walk(func(tmpl *template.ManagedTemplate) {
		fmt.Printf("%s: %s\n", tmpl.GetName(), tmpl.GetDescription())
	})

	if err != nil {
		cmd.Logger.WithError(err).Fatal("Error walking templates")
	}
}
