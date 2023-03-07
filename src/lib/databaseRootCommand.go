package lib

import (
	"context"
	"io"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

type dbCtx struct{}

type dbCmdConfig struct {
	OutF io.Writer
	ErrF io.Writer
	db   *Db
}

func NewDatabaseRootCmd(config *dbCmdConfig) *cobra.Command {
	var rootCmd = &cobra.Command{
		SilenceErrors: true,
		Short:         "Database manager cli",
		Example:       ".tables to list tables\n.schema to list schemas",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ctx := context.WithValue(cmd.Context(), dbCtx{}, config)
			cmd.SetContext(ctx)
		},
	}

	rootCmd.AddCommand(tableCmd)
	rootCmd.AddCommand(schemaCmd)
	rootCmd.DisableSuggestions = true
	return rootCmd
}

func CreateNewDatabaseRootCmd(config *dbCmdConfig) *cobra.Command {
	return NewDatabaseRootCmd(config)
}
