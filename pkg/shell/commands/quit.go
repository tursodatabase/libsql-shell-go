package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var quitCmd = &cobra.Command{
	Use:   ".quit",
	Short: "Exit this program",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, ok := cmd.Context().Value(dbCtx{}).(*DbCmdConfig)
		if !ok {
			return fmt.Errorf("missing db connection")
		}

		config.SetInterruptShell()
		return nil
	},
}
