package commands

import (
	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   ".help",
	Short: `List of all available commands.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cmd.Parent().Help()
		if err != nil {
			return err
		}

		return nil
	},
}
