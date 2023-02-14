package main

import (
	"testing"

	"github.com/chiselstrike/libsql-shell/cmd"
	qt "github.com/frankban/quicktest"
)

func TestRootCommand_WhenAllFlagsAreProvided_ExpectSQLStatementsExecutedSuccessfully(t *testing.T) {
	c := qt.New(t)

	dbPath := c.TempDir() + `\test.sqlite`
	rootCmd := cmd.NewRootCmd()

	out, err := execute(t, rootCmd, "--exec", "CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT);", "--db", dbPath)

	c.Assert(err, qt.IsNil)
	c.Assert(out, qt.Equals, "SQL statements executed successfully")
}

func TestRootCommand_WhenAllRequiredFlagsAreMissing_ExpectErrorReturned(t *testing.T) {
	c := qt.New(t)

	rootCmd := cmd.NewRootCmd()

	_, err := execute(t, rootCmd)

	c.Assert(err.Error(), qt.Equals, `required flag(s) "db", "exec" not set`)
}

func TestRootCommand_WhenDbFlagIsMissing_ExpectErrorReturned(t *testing.T) {
	c := qt.New(t)

	rootCmd := cmd.NewRootCmd()

	_, err := execute(t, rootCmd, "--exec", "CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT);")

	c.Assert(err.Error(), qt.Equals, `required flag(s) "db" not set`)
}

func TestRootCommand_WhenExecFlagIsMissing_ExpectErrorReturned(t *testing.T) {
	c := qt.New(t)

	rootCmd := cmd.NewRootCmd()

	_, err := execute(t, rootCmd, "--db", "test.sqlite")

	c.Assert(err.Error(), qt.Equals, `required flag(s) "exec" not set`)
}
