package lib

import (
	"fmt"

	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   ".schema",
	Short: `Show table schemas.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, ok := cmd.Context().Value(dbCtx{}).(*dbCmdConfig)
		if !ok {
			return fmt.Errorf("missing db connection")
		}

		schemaStatement := `select sql from sqlite_schema
			where name not like 'sqlite_%'
			and name != '_litestream_seq'
			and name != '_litestream_lock'
			and name != 'libsql_wasm_func_table'
			order by name`

		config.db.ExecuteAndPrintStatements(schemaStatement, config.OutF, config.ErrF, true)

		return nil
	},
}
