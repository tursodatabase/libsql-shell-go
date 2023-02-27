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

	_, err := utils.Execute(t, rootCmd, "--exec", "CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT);", dbPath)

	c.Assert(err, qt.IsNil)
}

func TestRootCommandFlags_WhenDbIsMissing_ExpectErrorReturned(t *testing.T) {
	c := qt.New(t)

	rootCmd := cmd.NewRootCmd()

	_, err := utils.Execute(t, rootCmd, "--exec", "CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT);")

	c.Assert(err.Error(), qt.Equals, `accepts 1 arg(s), received 0`)
}

func TestRootCommandFlags_GivenEmptyStatements_ExpectErrorReturned(t *testing.T) {
	tc := utils.NewTestContext(t)

	_, err := tc.Execute("")

	tc.Assert(err.Error(), qt.IsNotNil)
}
