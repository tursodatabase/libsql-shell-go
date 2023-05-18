package shellcmd

import (
	"fmt"

	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
	"github.com/spf13/cobra"
)

var tableCmd = &cobra.Command{
	Use:   ".tables",
	Short: `List all existing tables in the database.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, ok := cmd.Context().Value(dbCtx{}).(*DbCmdConfig)
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

		return config.Db.ExecuteAndPrintStatements(tableStatement, config.OutF, true, enums.TABLE_MODE)
	},
}
