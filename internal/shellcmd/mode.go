package shellcmd

import (
	"fmt"

	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
	"github.com/spf13/cobra"
)

var modeCmd = &cobra.Command{
	Use:   ".mode MODE",
	Short: "Set output mode",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, ok := cmd.Context().Value(dbCtx{}).(*DbCmdConfig)
		if !ok {
			return fmt.Errorf("missing db connection")
		}
		mode := args[0]
		switch mode {
		case string(enums.TABLE_MODE):
			config.SetMode(enums.TABLE_MODE)
		case string(enums.CSV_MODE):
			config.SetMode(enums.CSV_MODE)
		case string(enums.JSON_MODE):
			config.SetMode(enums.JSON_MODE)
		default:
			return fmt.Errorf("unsupported mode")
		}
		return nil
	},
}
