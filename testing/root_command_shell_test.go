package main_test

import (
	"testing"

	"github.com/chiselstrike/libsql-shell/src/lib"
	"github.com/chiselstrike/libsql-shell/testing/utils"
	qt "github.com/frankban/quicktest"
)

func TestRootCommandShell_WhenNoCommandsAreProvided_ExpectShellExecutedWithoutError(t *testing.T) {
	tc := utils.NewTestContext(t)

	result, err := tc.ExecuteShell([]string{})

	tc.Assert(err, qt.IsNil)
	tc.Assert(result, qt.Equals, "")
}

func TestRootCommandShell_WhenExecuteInvalidStatement_ExpectError(t *testing.T) {
	tc := utils.NewTestContext(t)

	result, err := tc.ExecuteShell([]string{"SELECTT 1;"})

	tc.Assert(err, qt.IsNil)
	tc.Assert(result, qt.Equals, "Error: near \"SELECTT\": syntax error")
}

func TestRootCommandShell_WhenCreateTable_ExpectDbHaveTheTable(t *testing.T) {
	tc := utils.NewTestContext(t)

	result, err := tc.ExecuteShell([]string{"CREATE TABLE test (name STRING);", "SELECT * FROM test;"})

	tc.Assert(err, qt.IsNil)
	tc.Assert(result, qt.Equals, "name")
}

func TestRootCommandShell_WhenCreateTableAndInsertData_ExpectDbHaveTheTableWithTheData(t *testing.T) {
	tc := utils.NewTestContext(t)

	result, err := tc.ExecuteShell([]string{"CREATE TABLE test (name STRING);", "INSERT INTO test VALUES ('test');", "SELECT * FROM test;"})

	tc.Assert(err, qt.IsNil)
	tc.Assert(result, qt.Equals, "name\ntest")
}

func TestRootCommandShell_WhenTypingQuitCommand_ExpectShellExit(t *testing.T) {
	tc := utils.NewTestContext(t)

	result, err := tc.ExecuteShell([]string{lib.QUIT_COMMAND, "SELECT 1;"})

	tc.Assert(err, qt.IsNil)
	tc.Assert(result, qt.Equals, "")
}
