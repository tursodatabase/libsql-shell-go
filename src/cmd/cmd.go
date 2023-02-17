package cmd

import (
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	"github.com/chiselstrike/libsql-shell/src/lib"
)

type RootArgs struct {
	statements string
}

func NewRootCmd() *cobra.Command {
	var rootArgs RootArgs = RootArgs{}
	var rootCmd = &cobra.Command{
		SilenceUsage: true,
		Use:          "libsql-shell",
		Short:        "A cli for executing SQL statements on a libSQL or SQLite database",
		Args:         cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath := args[0]
			db, err := lib.NewSQLite3(dbPath)
			if err != nil {
				return err
			}
			defer db.Close()

			if len(rootArgs.statements) == 0 {
				l, err := lib.NewReadline(cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr())
				if err != nil {
					return err
				}
				return db.RunShell(l)
			}

			result, err := db.ExecuteStatements(rootArgs.statements)
			if err != nil {
				return err
			}
			cmd.Println(result)
			return nil
		},
	}

	rootCmd.Flags().StringVarP(&rootArgs.statements, "exec", "e", "", "SQL statements separated by ;")

	return rootCmd
}

func Init() {
	var rootCmd *cobra.Command = NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
