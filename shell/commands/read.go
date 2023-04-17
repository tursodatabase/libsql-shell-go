package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   ".read FILENAME",
	Short: "Execute commands from a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, ok := cmd.Context().Value(dbCtx{}).(*DbCmdConfig)
		if !ok {
			return fmt.Errorf("missing db connection")
		}

		file, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			return err
		}

		content := make([]byte, stat.Size())
		_, err = file.Read(content)
		if err != nil {
			return err
		}

		return config.Db.ExecuteAndPrintStatements(strings.TrimSpace(string(content)), config.OutF, false, config.GetMode())
	},
}
