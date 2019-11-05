package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/bserdar/watermelon/server"
	"github.com/bserdar/watermelon/server/inventory"
	"github.com/bserdar/watermelon/server/inventory/yml"
)

var verbose bool
var inventoryFile string

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose")
	rootCmd.PersistentFlags().StringVarP(&inventoryFile, "inv", "i", "", "Inventory file (YAML)")
}

func loadInventory(session server.Session) (*inventory.InvServer, map[string]interface{}) {
	if len(inventoryFile) == 0 {
		return inventory.NewInvServer([]*server.Host{}, server.InventoryConfiguration{}, session), nil
	}
	log.Debugf("Loading inventory %s", inventoryFile)

	cfg, hosts, config, err := yml.LoadInventory(inventoryFile)
	if err != nil {
		panic(err)
	}
	log.Debugf("There are %d hosts", len(hosts))
	return inventory.NewInvServer(hosts, cfg, session), config
}

// rootcmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wm",
	Short: "Watermelon: Imperative configuration automation engine.",
	Long:  `Imperative Configuration automation engine.`}

// Execute the root cmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
