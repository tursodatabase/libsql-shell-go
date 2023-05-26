package main_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/stretchr/testify/suite"

	"github.com/libsql/libsql-shell-go/test/utils"
)

type RootCommandShellSuite struct {
	suite.Suite

	dbPath string
	tc     *utils.DbTestContext
}

func NewRootCommandShellSuite(dbPath string) *RootCommandShellSuite {
	return &RootCommandShellSuite{dbPath: dbPath}
}

func (s *RootCommandShellSuite) SetupSuite() {
	s.tc = utils.NewTestContext(s.T(), s.dbPath)
	s.tc.DropAllTables()
}

func (s *RootCommandShellSuite) TearDownSuite() {
	s.tc.Close()
}

func (s *RootCommandShellSuite) TearDownTest() {
	s.tc.DropAllTables()
}

func (s *RootCommandShellSuite) Test_WhenCreateTable_ExpectDbHaveTheTable() {
	outS, errS, err := s.tc.ExecuteShell([]string{"CREATE TABLE test (name STRING);", "SELECT * FROM test;"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{"name"}, [][]string{}))
}

func (s *RootCommandShellSuite) Test_WhenCreateTableAndInsertData_ExpectDbHaveTheTableWithTheData() {
	outS, errS, err := s.tc.ExecuteShell([]string{"CREATE TABLE test (name STRING);", "INSERT INTO test VALUES ('test');", "SELECT * FROM test;"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{"name"}, [][]string{{"test"}}))
}

func (s *RootCommandShellSuite) Test_WhenNoCommandsAreProvided_ExpectShellExecutedWithoutError() {
	outS, errS, err := s.tc.ExecuteShell([]string{})

	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outS, qt.Equals, "")
}

func (s *RootCommandShellSuite) Test_WhenExecuteInvalidStatement_ExpectError() {
	outS, errS, err := s.tc.ExecuteShell([]string{"SELECTT 1;"})
	s.tc.Assert(err, qt.IsNil)

	s.tc.Assert(outS, qt.Equals, "")
	s.tc.Assert(len(errS), qt.Not(qt.Equals), 0)
}

func (s *RootCommandShellSuite) Test_WhenTypingQuitCommand_ExpectShellNotRunFollowingCommands() {
	outS, errS, err := s.tc.ExecuteShell([]string{".quit", "SELECT 1;"})

	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(outS, qt.Equals, "")
	s.tc.Assert(errS, qt.Equals, "")
}

func (s *RootCommandShellSuite) TestRootCommandShell_WhenSplittingStatementsIntoMultipleLine_ExpectMergeLinesBeforeExecuting() {
	s.tc.CreateSimpleTable("simple_table", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}})

	outS, errS, err := s.tc.ExecuteShell([]string{"SELECT", "*", "FROM", "simple_table;"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{"id", "textField", "intField"}, [][]string{{"1", "value1", "1"}}))
}

func TestRootCommandShellSuite_WhenDbIsSQLite(t *testing.T) {
	suite.Run(t, NewRootCommandShellSuite(t.TempDir()+"test.sqlite"))
}

func TestRootCommandShellSuite_WhenDbIsSqld(t *testing.T) {
	testConfig := utils.GetTestConfig(t)
	if testConfig.SkipSqldTests {
		t.Skip("Skipping Sqld tests due configuration")
	}

	suite.Run(t, NewRootCommandShellSuite(testConfig.SqldDbPath))
}
