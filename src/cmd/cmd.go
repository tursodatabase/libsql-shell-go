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
	dbPath     string
}

func NewRootCmd() *cobra.Command {
	var rootArgs RootArgs = RootArgs{}
	var rootCmd = &cobra.Command{
		Use:   "libsql-shell",
		Short: "A cli for executing SQL statements on a libSQL or SQLite database",
		Run: func(cmd *cobra.Command, args []string) {
			db, err := lib.OpenOrCreateSQLite3(rootArgs.dbPath)
			if err != nil {
				cmd.PrintErrln("Error opening database:", err)
				os.Exit(1)
			}
			defer db.Close()

			result, err := lib.ExecuteStatements(db, rootArgs.statements)
			if err != nil {
				cmd.PrintErrln("Error executing SQL statements:", err)
				os.Exit(1)
			}
			cmd.Println(result)
		},
	}

	rootCmd.Flags().StringVarP(&rootArgs.statements, "exec", "e", "", "SQL statements separated by ;")
	rootCmd.MarkFlagRequired("exec")
	rootCmd.Flags().StringVarP(&rootArgs.dbPath, "db", "d", "", "Path to database file")
	rootCmd.MarkFlagRequired("db")

	return rootCmd
}

func Init() {
	var rootCmd *cobra.Command = NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}