package main_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/stretchr/testify/suite"

	"github.com/libsql/libsql-shell-go/test/utils"
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

func (s *DBRootCommandShellSuite) Test_GivenADBWithTwoTables_WhenCallDotSchemaCommandWithPattern_ExpectToReturnOneSchema() {
	s.tc.CreateEmptySimpleTable("simple_table")
	s.tc.CreateEmptySimpleTable("another_simple_table")

	outSchema, errS, err := s.tc.ExecuteShell([]string{".schema simple_table"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outSchema, qt.Equals, utils.GetPrintTableOutput([]string{""}, [][]string{{"CREATE TABLE simple_table (id INTEGER PRIMARY KEY, textField TEXT, intField INTEGER)"}}))
}

func (s *DBRootCommandShellSuite) Test_GivenADBWithThreeTables_WhenCallDotSchemaCommandWithPartialDbName_ExpectToReturnTwoSchemas() {
	s.tc.CreateEmptySimpleTable("simple_table")
	s.tc.CreateEmptySimpleTable("test_table_one")
	s.tc.CreateEmptySimpleTable("test_table_two")

	outSchema, errS, err := s.tc.ExecuteShell([]string{".schema test"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outSchema, qt.Equals, utils.GetPrintTableOutput([]string{""}, [][]string{{"CREATE TABLE test_table_one (id INTEGER PRIMARY KEY, textField TEXT, intField INTEGER)\nCREATE TABLE test_table_two (id INTEGER PRIMARY KEY, textField TEXT, intField INTEGER)"}}))
}

func (s *DBRootCommandShellSuite) Test_GivenADBWithTwoTables_WhenCallDotSchemaCommandWithPatternThatDoesNotMatch_ExpectEmptyReturn() {
	s.tc.CreateEmptySimpleTable("simple_table")
	s.tc.CreateEmptySimpleTable("another_simple_table")

	outSchema, errS, err := s.tc.ExecuteShell([]string{".schema non_existing_table"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outSchema, qt.Equals, utils.GetPrintTableOutput([]string{""}, [][]string{{""}}))
}

func (s *DBRootCommandShellSuite) Test_WhenCallDotHelpCommand_ExpectAListWithAllAvailableCommands() {
	outS, errS, err := s.tc.ExecuteShell([]string{".help"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outS, qt.Equals, ".help       List of all available commands.\n  .read       Execute commands from a file\n  .schema     Show table schemas.\n  .tables     List all existing tables in the database.\n  .quit       Exit this program.")
}

func (s *DBRootCommandShellSuite) Test_GivenAEmptyDb_WhenCallDotReadCommand_ExpectToSeeATableWithOneEntry() {
	content := `CREATE TABLE IF NOT EXISTS testread (name TEXT);
		/* Comment in the middle of the file.*/
		INSERT INTO testread VALUES("test");
		
		SELECT * FROM testread;`
	file, filePath := s.tc.CreateTempFile(content)

	defer file.Close()

	outS, errS, err := s.tc.ExecuteShell([]string{".read " + filePath})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{"NAME"}, [][]string{{"test"}}))
}

func (s *DBRootCommandShellSuite) Test_GivenAEmptyDb_WhenCallDotReadCommandPassingANonExistingFile_ExpectToReturnAnErrorMessage() {
	outS, errS, err := s.tc.ExecuteShell([]string{".read nonExistingFile.txt"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "Error: open nonExistingFile.txt: no such file or directory")
	s.tc.Assert(outS, qt.Equals, "")
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
