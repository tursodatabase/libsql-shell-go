package shell

import (
	"fmt"

	"github.com/spf13/cobra"
)

var indexesCmd = &cobra.Command{
	Use:   ".indexes ?TABLE?",
	Short: "List indexes in a table or database",
	Long:  `List all indexes in a table or in the entire database if no table is specified.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, ok := cmd.Context().Value(dbCtx{}).(*dbCmdConfig)
		if !ok {
			return fmt.Errorf("missing db connection")
		}
		var schemaStatement string

		if len(args) == 1 {
			schemaStatement = "SELECT name FROM sqlite_master WHERE type='index' AND tbl_name like '" + args[0] + "'"
		} else {
			schemaStatement = "SELECT name FROM sqlite_master WHERE type='index'"
		}

		config.db.ExecuteAndPrintStatements(schemaStatement, config.OutF, config.ErrF, true)

		return nil
	},
}
