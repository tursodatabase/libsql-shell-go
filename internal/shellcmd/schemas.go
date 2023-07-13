package shellcmd

import (
	"fmt"

	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   ".schema ?PATTERN?",
	Short: `Show table schemas.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, ok := cmd.Context().Value(dbCtx{}).(*DbCmdConfig)
		if !ok {
			return fmt.Errorf("missing db connection")
		}

		schemaStatement := `select sql || ';' from sqlite_schema
			where name not like 'sqlite_%'
			and name != '_litestream_seq'
			and name != '_litestream_lock'
			and name != 'libsql_wasm_func_table'`

		if len(args) == 1 {
			schemaStatement += " and name like '" + args[0] + "'"
		}

		schemaStatement += " order by name"

		return config.Db.ExecuteAndPrintStatements(schemaStatement, config.OutF, true, enums.TABLE_MODE)
	},
}
