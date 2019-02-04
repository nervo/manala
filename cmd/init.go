package cmd

import (
	"fmt"
	"github.com/apex/log"
	"github.com/fgrosse/goldi"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"manala/pkg/project"
	"manala/pkg/syncer"
	"manala/pkg/template"
	"strings"
)

/*********/
/* Cobra */
/*********/

func InitCobra(container *goldi.Container) *cobra.Command {

	var opt InitOptions

	cmd := &cobra.Command{
		Use:     "init [DIR]",
		Aliases: []string{"in"},
		Short:   "Init project",
		Long: `Init (manala init) will init project.

A optional dir could be passed as argument.

Example: manala init -> resulting in an init in current directory
Example: manala init /foo/bar -> resulting in an init in /foo/bar directory`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				args = append(args, "")
			}
			container.MustGet("cmd.init").(*InitCmd).Run(args[0], opt)
		},
	}

	return cmd
}

/***********/
/* Options */
/***********/

type InitOptions struct {
}

/***********/
/* Command */
/***********/

type InitCmd struct {
	ProjectManager  project.ManagerInterface
	TemplateManager template.ManagerInterface
	Syncer          syncer.Interface
	Logger          log.Interface
}

func (cmd *InitCmd) Run(dir string, opt InitOptions) {
	// Get real directory
	dir, err := getRealDir(dir)
	if err != nil {
		cmd.Logger.WithError(err).Fatal("Error getting real directory")
	}

	// Check project already initialized at directory
	_, err = cmd.ProjectManager.Get(dir)
	if err == nil {
		cmd.Logger.WithField("dir", dir).Fatal("Project already initialized")
	}

	var templates []template.Interface

	// Walk into templates
	err = cmd.TemplateManager.Walk(func(tmpl *template.ManagedTemplate) {
		templates = append(templates, tmpl)
	})

	if err != nil {
		cmd.Logger.WithError(err).Fatal("Error walking templates")
	}

	prompt := promptui.Select{
		Items: templates,
		Templates: &promptui.SelectTemplates{
			Label:    "Select template:",
			Active:   `{{ "▸" | bold }} {{ .GetName | underline }}`,
			Inactive: "  {{ .GetName }}",
			Selected: `{{ "✔" | green }} {{ .GetName | faint }}`,
			Details: `
{{ .GetDescription }}`,
		},
		Searcher: func(input string, index int) bool {
			template := templates[index]
			name := strings.Replace(strings.ToLower(template.GetName()), " ", "", -1)
			description := strings.Replace(strings.ToLower(template.GetDescription()), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input) || strings.Contains(description, input)
		},
		Size:              12,
		StartInSearchMode: true,
	}

	i, _, err := prompt.Run()

	if err != nil {
		switch err {
		case promptui.ErrInterrupt:
			cmd.Logger.Fatal("Interruption")
		default:
			cmd.Logger.WithError(err).Fatal("Error prompting")
		}
	}

	fmt.Printf("You choose template %d: %s\n", i+1, templates[i].GetName())
}
