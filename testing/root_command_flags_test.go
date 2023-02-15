package main_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/chiselstrike/libsql-shell/src/cmd"
	"github.com/chiselstrike/libsql-shell/testing/utils"
)

func TestRootCommandFlags_WhenAllFlagsAreProvided_ExpectSQLStatementsExecutedWithoutError(t *testing.T) {
	c := qt.New(t)

	dbPath := c.TempDir() + `\test.sqlite`
	rootCmd := cmd.NewRootCmd()

	_, err := utils.Execute(t, rootCmd, "--exec", "CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT);", "--db", dbPath)

	c.Assert(err, qt.IsNil)
}

func TestRootCommandFlags_WhenAllRequiredFlagsAreMissing_ExpectErrorReturned(t *testing.T) {
	c := qt.New(t)

	rootCmd := cmd.NewRootCmd()

	_, err := utils.Execute(t, rootCmd)

	c.Assert(err.Error(), qt.Equals, `required flag(s) "db", "exec" not set`)
}

func TestRootCommandFlags_WhenDbFlagIsMissing_ExpectErrorReturned(t *testing.T) {
	c := qt.New(t)

	rootCmd := cmd.NewRootCmd()

	_, err := utils.Execute(t, rootCmd, "--exec", "CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT);")

	c.Assert(err.Error(), qt.Equals, `required flag(s) "db" not set`)
}

func TestRootCommandFlags_WhenExecFlagIsMissing_ExpectErrorReturned(t *testing.T) {
	c := qt.New(t)

	rootCmd := cmd.NewRootCmd()

	_, err := utils.Execute(t, rootCmd, "--db", "test.sqlite")

	c.Assert(err.Error(), qt.Equals, `required flag(s) "exec" not set`)
}
