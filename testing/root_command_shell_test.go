package main_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/stretchr/testify/suite"

	"github.com/chiselstrike/libsql-shell/src/lib"
	"github.com/chiselstrike/libsql-shell/testing/utils"
)

type RootCommandShellSuite struct {
	suite.Suite

	// dbInTest utils.DbType
	dbPath string
	tc     *utils.DbTestContext
}

func NewRootCommandShellSuite(dbPath string) *RootCommandShellSuite {
	return &RootCommandShellSuite{dbPath: dbPath}
}

func (s *RootCommandShellSuite) SetupTest() {
	s.tc = utils.NewTestContext(s.T(), s.dbPath)
}

func (s *RootCommandShellSuite) TearDownTest() {
	s.tc.TearDown()
}

func (s *RootCommandShellSuite) TestRootCommandShell_WhenCreateTable_ExpectDbHaveTheTable() {
	result, err := s.tc.ExecuteShell([]string{"CREATE TABLE test (name STRING);", "SELECT * FROM test;"})

	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(result, qt.Equals, utils.GetPrintTableOutput([]string{"name"}, [][]string{}))
}

func (s *RootCommandShellSuite) TestRootCommandShell_WhenCreateTableAndInsertData_ExpectDbHaveTheTableWithTheData() {
	result, err := s.tc.ExecuteShell([]string{"CREATE TABLE test (name STRING);", "INSERT INTO test VALUES ('test');", "SELECT * FROM test;"})

	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(result, qt.Equals, utils.GetPrintTableOutput([]string{"name"}, [][]string{{"test"}}))
}

func (s *RootCommandShellSuite) TestRootCommandShell_WhenNoCommandsAreProvided_ExpectShellExecutedWithoutError() {
	result, err := s.tc.ExecuteShell([]string{})

	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(result, qt.Equals, "")
}

func (s *RootCommandShellSuite) TestRootCommandShell_WhenExecuteInvalidStatement_ExpectError() {
	result, err := s.tc.ExecuteShell([]string{"SELECTT 1;"})

	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(result, qt.Equals, "Error: near \"SELECTT\": syntax error")
}

func (s *RootCommandShellSuite) TestRootCommandShell_WhenTypingQuitCommand_ExpectShellNotRunFollowingCommands() {
	result, err := s.tc.ExecuteShell([]string{lib.QUIT_COMMAND, "SELECT 1;"})

	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(result, qt.Equals, "")
}

func TestRootCommandShellSuite_WhenDbIsSQLite(t *testing.T) {
	suite.Run(t, NewRootCommandShellSuite(t.TempDir()+"test.sqlite"))
}
