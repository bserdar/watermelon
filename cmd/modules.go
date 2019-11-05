package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bserdar/watermelon/server"
)

func init() {
	rootCmd.AddCommand(modulesCmd)
}

// modulesCmd returns information about modules
var modulesCmd = &cobra.Command{
	Use:   "describe",
	Short: "List available modules, or get module description.",
	Long:  `List available modules, or get module description.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			for name, mod := range server.LocalModules {
				fmt.Printf("%s\t%s\n", name, mod.Describe())
			}
		} else {
			mod, ok := server.LocalModules[args[0]]
			if !ok {
				fmt.Printf("Cannot find module %s\n", args[0])
				return
			}
			fmt.Printf(mod.Help())
		}
	}}
