package shellcmd

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/libsql/libsql-shell-go/internal/db"
	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
)

type dbCtx struct{}

type DbCmdConfig struct {
	OutF              io.Writer
	ErrF              io.Writer
	Db                *db.Db
	SetInterruptShell func()
	SetMode           func(mode enums.PrintMode)
	GetMode           func() enums.PrintMode
}

const helpTemplate = `{{range .Commands}}{{if (and (not .Hidden) (or .IsAvailableCommand) (ne .Name "completion"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}
`

func NewDatabaseRootCmd(config *DbCmdConfig) *cobra.Command {
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

	rootCmd.AddCommand(tableCmd, schemaCmd, helpCmd, readCmd, indexesCmd, quitCmd, dumpCmd, modeCmd)
	rootCmd.SetOut(config.OutF)
	rootCmd.SetErr(config.ErrF)
	rootCmd.SetHelpTemplate(helpTemplate)
	return rootCmd
}

func CreateNewDatabaseRootCmd(config *DbCmdConfig) *cobra.Command {
	return NewDatabaseRootCmd(config)
}
