package main_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/stretchr/testify/suite"

	"github.com/chiselstrike/libsql-shell/testing/utils"
)

type DBRootCommandShellSuite struct {
	suite.Suite

	dbPath string
	tc     *utils.DbTestContext
}

func NewDBRootCommandShellSuite(dbPath string) *DBRootCommandShellSuite {
	return &DBRootCommandShellSuite{dbPath: dbPath}
}

func (s *DBRootCommandShellSuite) SetupSuite() {
	s.tc = utils.NewTestContext(s.T(), s.dbPath)
	s.tc.DropAllTables()
}

func (s *DBRootCommandShellSuite) TearDownTest() {
	s.tc.DropAllTables()
}

func (s *DBRootCommandShellSuite) Test_GivenADBWithTwoTables_WhenCallDotTablesCommand_ExpectAListContainingTheTableNames() {
	s.tc.CreateEmptySimpleTable("simple_table")
	s.tc.CreateEmptySimpleTable("another_simple_table")

	outTables, errS, err := s.tc.ExecuteShell([]string{".tables"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outTables, qt.Equals, utils.GetPrintTableOutput([]string{""}, [][]string{{"another_simple_table\nsimple_table"}}))
}

func (s *DBRootCommandShellSuite) Test_GivenADBWithTwoTables_WhenCallDotSchemaCommand_ExpectAListContainingTheSchemas() {
	s.tc.CreateEmptySimpleTable("simple_table")
	s.tc.CreateEmptySimpleTable("another_simple_table")

	outSchema, errS, err := s.tc.ExecuteShell([]string{".schema"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outSchema, qt.Equals, utils.GetPrintTableOutput([]string{""}, [][]string{{"CREATE TABLE another_simple_table (id INTEGER PRIMARY KEY, textField TEXT, intField INTEGER)\nCREATE TABLE simple_table (id INTEGER PRIMARY KEY, textField TEXT, intField INTEGER)"}}))
}

func (s *DBRootCommandShellSuite) Test_WhenCallACommandThatDoesNotExist_ExpectToReturnAnErrorMessage() {
	outS, errS, err := s.tc.ExecuteShell([]string{".nonExistingCommand"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, `Error: unknown command or invalid arguments: ".nonExistingCommand". Enter ".help" for help`)
	s.tc.Assert(outS, qt.Equals, "")
}

func TestDBRootCommandShellSuite_WhenDbIsSQLite(t *testing.T) {
	suite.Run(t, NewDBRootCommandShellSuite(t.TempDir()+"test.sqlite"))
}

func TestDBRootCommandShellSuite_WhenDbIsTurso(t *testing.T) {
	testConfig := utils.GetTestConfig(t)
	if testConfig.SkipTursoTests {
		t.Skip("Skipping Turso tests due configuration")
	}

	suite.Run(t, NewDBRootCommandShellSuite(testConfig.TursoDbPath))
}
