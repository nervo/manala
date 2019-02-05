package cmd

import (
	"fmt"
	"github.com/apex/log"
	"github.com/fgrosse/goldi"
	"github.com/spf13/cobra"
	"manala/pkg/template"
	"manala/pkg2/zroject"
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

	//path := "/Volumes/Data/manala/manala.test/project/.manala.yaml"
	//path := "/home/nervo/workspace/manala/manala.test/project/.manala.yaml"

	//dir := "/Volumes/Data/manala/manala.test/project";
	dir := "/home/nervo/workspace/manala/manala.test/project"

	/*
		handler := zroject.NewYamlConfigHandler(path)

		go func() {
			for {
				select {
				case event, ok := <-handler.Events:
					fmt.Printf("%#v\n", event)
					if !ok {
						fmt.Printf("Pas ok...\n")
						return
					}
				}
			}
		}()

		cfg, err := handler.Load()
		if err != nil {
			cmd.Logger.WithError(err).Fatal("Error loading config")
		}

		fmt.Printf("%#v\n", cfg)
	*/

	//projectHandler := zroject.NewDirZrojectHandler(dir)

	//project, err := projectHandler.Load()
	project, err := zroject.Load(dir)
	if err != nil {
		cmd.Logger.WithError(err).Fatal("Error loading project")
	}

	fmt.Printf("%#v\n", project)

	/*
		fmt.Printf("%#v\n", project.GetFs())

		options, _ := project.GetOptions()
		fmt.Printf("%#v\n", options)

		options, _ = project.GetOptions("dist")
		fmt.Printf("%#v\n", options)

		options, _ = project.GetOptions("local")
		fmt.Printf("%#v\n", options)

		options, _ = project.GetOptions("local", "dist")
		fmt.Printf("%#v\n", options)

		options, _ = project.GetOptions("dist", "local")
		fmt.Printf("%#v\n", options)

		options, err = project.GetOptions("prout")
		if err != nil {
			cmd.Logger.WithError(err).Fatal("Error getting options")
		}
	*/

	/*
		err = handler.Dump(cfg)
		if err != nil {
			cmd.Logger.WithError(err).Fatal("Error dumping config")
		}
	*/

	/*
		// Walk into templates
		err := cmd.TemplateManager.Walk(func(tmpl *template.ManagedTemplate) {
			fmt.Printf("%s: %s\n", tmpl.GetName(), tmpl.GetDescription())
		})

		if err != nil {
			cmd.Logger.WithError(err).Fatal("Error walking templates")
		}
	*/
}
