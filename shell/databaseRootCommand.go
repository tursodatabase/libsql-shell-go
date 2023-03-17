package shell

import (
	"context"
	"io"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	"github.com/libsql/libsql-shell-go/pkg/libsql"
)

type dbCtx struct{}

type dbCmdConfig struct {
	OutF io.Writer
	ErrF io.Writer
	db   *libsql.Db
}

const helpTemplate = `{{range .Commands}}{{if (and (not .Hidden) (or .IsAvailableCommand) (ne .Name "completion"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}
  {{rpad ".quit" .NamePadding }} Exit this program.
`

func NewDatabaseRootCmd(config *dbCmdConfig) *cobra.Command {
	var rootCmd = &cobra.Command{
		SilenceUsage:       true,
		SilenceErrors:      true,
		DisableSuggestions: true,
		Short:              "Database manager cli",
		Example:            ".tables to list tables\n.schema to list schemas",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ctx := context.WithValue(cmd.Context(), dbCtx{}, config)
			cmd.SetContext(ctx)
		},
	}

	rootCmd.AddCommand(tableCmd, schemaCmd, helpCmd, readCmd, indexesCmd)
	rootCmd.SetOut(config.OutF)
	rootCmd.SetErr(config.ErrF)
	rootCmd.SetHelpTemplate(helpTemplate)
	return rootCmd
}

func CreateNewDatabaseRootCmd(config *dbCmdConfig) *cobra.Command {
	return NewDatabaseRootCmd(config)
}
