package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ovh/venom"
)

// Cmd version
var Cmd = &cobra.Command{
	Use:     "version",
	Short:   "Display Version of venom: venom version",
	Long:    `venom version`,
	Aliases: []string{"v"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version venom: %s", venom.Version)
		if venom.BuildTime != "" {
			fmt.Printf(" (built: %s)", venom.BuildTime)
		}
		fmt.Println()
	},
}
