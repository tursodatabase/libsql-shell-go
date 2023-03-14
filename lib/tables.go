package lib

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tableCmd = &cobra.Command{
	Use:   ".tables",
	Short: `List all existing tables in the database.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, ok := cmd.Context().Value(dbCtx{}).(*dbCmdConfig)
		if !ok {
			return fmt.Errorf("missing db connection")
		}

		tableStatement := `select name from sqlite_schema
			where type = 'table'
			and name not like 'sqlite_%'
			and name != '_litestream_seq'
			and name != '_litestream_lock'
			and name != 'libsql_wasm_func_table'
			order by name`

		config.db.ExecuteAndPrintStatements(tableStatement, config.OutF, config.ErrF, true)

		return nil
	},
}
