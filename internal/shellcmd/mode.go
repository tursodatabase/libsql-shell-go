package shellcmd

import (
	"fmt"
	"strings"

	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
	"github.com/spf13/cobra"
)

var modeCmd = &cobra.Command{
	Use:   ".mode MODE",
	Short: "Set output mode",
	Args:  cobra.MaximumNArgs(1),
	ValidArgs: []string{
		string(enums.TABLE_MODE),
		string(enums.JSON_MODE),
		string(enums.CSV_MODE),
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		validModes := strings.Join(cmd.ValidArgs, ", ")
		config, ok := cmd.Context().Value(dbCtx{}).(*DbCmdConfig)
		if !ok {
			return fmt.Errorf("missing db connection")
		}
		currentMode := config.GetMode()
		if len(args) == 0 {
			return fmt.Errorf("No mode provided. Current mode is %s. Valid modes are %s", currentMode, validModes)
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
			return fmt.Errorf("Invalid mode. Current mode is %s. Valid modes are %s", currentMode, validModes)
		}
		return nil
	},
}
