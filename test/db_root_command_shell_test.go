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

	outSchema, errS, err := s.tc.ExecuteShell([]string{".schema test%"})
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

	expectedHelp :=
		`.dump       Render database content as SQL
  .help       List of all available commands.
  .indexes    List indexes in a table or database
  .mode       Set output mode
  .quit       Exit this program
  .read       Execute commands from a file
  .schema     Show table schemas.
  .tables     List all existing tables in the database.`
	s.tc.Assert(outS, qt.Equals, expectedHelp)
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

func (s *DBRootCommandShellSuite) Test_GivenADBWithTwoTables_WhenCreateTwoIndexesAndCallDotIndexesCommand_ExpectToReturnTheIndexes() {
	s.tc.CreateSimpleTable("simple_table", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})
	s.tc.CreateSimpleTable("simple_table_2", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})

	_, errS, err := s.tc.Execute("CREATE INDEX idx_textfield on simple_table (TextField);CREATE INDEX idx_intfield on simple_table_2 (IntField)")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	outS, errS, err := s.tc.ExecuteShell([]string{".indexes"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{""}, [][]string{{"idx_textfield\nidx_intfield"}}))
}

func (s *DBRootCommandShellSuite) Test_GivenADBWithThreeTables_WhenCreateThreeIndexesAndCallDotIndexesCommandPassingExactTableName_ExpectToReturnJustOneIndex() {
	s.tc.CreateSimpleTable("simple_table", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})
	s.tc.CreateSimpleTable("simple_table_2", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})
	s.tc.CreateSimpleTable("simple_table_3", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})

	_, errS, err := s.tc.Execute("CREATE INDEX idx_textfield on simple_table (TextField);CREATE INDEX idx_intfield on simple_table_2 (IntField);CREATE INDEX idx_intfield_third_table on simple_table_3 (IntField)")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	outS, errS, err := s.tc.ExecuteShell([]string{".indexes simple_table_3"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{""}, [][]string{{"idx_intfield_third_table"}}))
}

func (s *DBRootCommandShellSuite) Test_GivenADBWithThreeTables_WhenCreateThreeIndexesAndCallDotIndexesCommandPassingPartOfTableName_ExpectToReturnTwoIndexes() {
	s.tc.CreateSimpleTable("simple_table", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})
	s.tc.CreateSimpleTable("simple_table_2", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})
	s.tc.CreateSimpleTable("another_simple_table", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})

	_, errS, err := s.tc.Execute("CREATE INDEX idx_textfield on simple_table (TextField);CREATE INDEX idx_intfield on simple_table_2 (IntField);CREATE INDEX idx_intfield_another_table on another_simple_table (IntField)")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	outS, errS, err := s.tc.ExecuteShell([]string{".indexes simple%"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{""}, [][]string{{"idx_textfield\nidx_intfield"}}))
}

func (s *DBRootCommandShellSuite) Test_GivenADBWithATable_WhenCreateAIndexAndCallDotIndexesCommandPassingAWrongTableName_ExpectEmptyReturn() {
	s.tc.CreateSimpleTable("simple_table", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})

	_, errS, err := s.tc.Execute("CREATE INDEX idx_textfield on simple_table (TextField);")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	outS, errS, err := s.tc.ExecuteShell([]string{".indexes nonExistingTable"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{""}, [][]string{{""}}))
}

func (s *DBRootCommandShellSuite) Test_GivenAEmptyTable_WhenCallDotDumpCommand_ExpectNoErrors() {
	s.tc.CreateEmptyAllTypesTable("alltypes")

	outS, errS, err := s.tc.ExecuteShell([]string{".dump"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, "PRAGMA foreign_keys=OFF;\n"+
		`CREATE TABLE alltypes(textNullable TEXT, textNotNullable TEXT NOT NULL, textWithDefault TEXT DEFAULT 'defaultValue', intNullable INTEGER, intNotNullable INTEGER NOT NULL, intWithDefault INTEGER DEFAULT '0', floatNullable REAL, floatNotNullable REAL NOT NULL, floatWithDefault REAL DEFAULT '0.0', unknownNullable NUMERIC, unknownNotNullable NUMERIC NOT NULL, unknownWithDefault NUMERIC DEFAULT 0.0, blobNullable BLOB, blobNotNullable BLOB NOT NULL, blobWithDefault BLOB DEFAULT 'x"0"');`)
}

func (s *DBRootCommandShellSuite) Test_GivenATableConainingRandomFields_WhenInsertAndCallDotDumpCommand_ExpectNoErrors() {
	_, errS, err := s.tc.Execute(`CREATE TABLE alltypes (t text, i integer, r real, b blob);
	INSERT INTO alltypes VALUES ('text', 99, 3.14, x'0123456789ABCDEF')`)
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	outS, errS, err := s.tc.ExecuteShell([]string{".dump"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, "PRAGMA foreign_keys=OFF;\nCREATE TABLE alltypes(t TEXT, i INTEGER, r REAL, b BLOB);\nINSERT INTO alltypes VALUES ('text', 99, 3.14, X'0123456789ABCDEF');")
}

func (s *DBRootCommandShellSuite) Test_GivenATableConainingFieldsWithALLTypes_WhenInsertAndCallDotDumpCommand_ExpectNoErrors() {
	s.tc.CreateAllTypesTable("alltypes", []utils.AllTypesTableEntry{
		{TextNotNullable: "text2", IntNotNullable: 0, FloatNotNullable: 1.5, UnknownNotNullable: 0.0, BlobNotNullable: "0123456789ABCDEF"},
	})

	outS, errS, err := s.tc.ExecuteShell([]string{".dump"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, "PRAGMA foreign_keys=OFF;\n"+
		`CREATE TABLE alltypes(textNullable TEXT, textNotNullable TEXT NOT NULL, textWithDefault TEXT DEFAULT 'defaultValue', intNullable INTEGER, intNotNullable INTEGER NOT NULL, intWithDefault INTEGER DEFAULT '0', floatNullable REAL, floatNotNullable REAL NOT NULL, floatWithDefault REAL DEFAULT '0.0', unknownNullable NUMERIC, unknownNotNullable NUMERIC NOT NULL, unknownWithDefault NUMERIC DEFAULT 0.0, blobNullable BLOB, blobNotNullable BLOB NOT NULL, blobWithDefault BLOB DEFAULT 'x"0"');`+
		"\n"+`INSERT INTO alltypes VALUES (NULL, 'text2', 'defaultValue', NULL, 0, 0, NULL, 1.5, 0, NULL, 0, 0, NULL, X'0123456789ABCDEF', 'x"0"');`)
}

func (s *DBRootCommandShellSuite) Test_GivenATableWithRecords_WhenCreateIndexAndCallDotDumpCommand_ExpectNoErrors() {
	s.tc.CreateEmptyAllTypesTable("alltypes")
	_, errS, err := s.tc.Execute("CREATE INDEX idx_textNullable on alltypes (textNullable);CREATE INDEX idx_intNotNullable on alltypes (intNotNullable) WHERE intNotNullable > 1;")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	outS, errS, err := s.tc.ExecuteShell([]string{".dump"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, "PRAGMA foreign_keys=OFF;\n"+
		"CREATE TABLE alltypes(textNullable TEXT, textNotNullable TEXT NOT NULL, textWithDefault TEXT DEFAULT 'defaultValue', intNullable INTEGER, intNotNullable INTEGER NOT NULL, intWithDefault INTEGER DEFAULT '0', floatNullable REAL, floatNotNullable REAL NOT NULL, floatWithDefault REAL DEFAULT '0.0', unknownNullable NUMERIC, unknownNotNullable NUMERIC NOT NULL, unknownWithDefault NUMERIC DEFAULT 0.0, blobNullable BLOB, blobNotNullable BLOB NOT NULL, blobWithDefault BLOB DEFAULT 'x\"0\"');\n"+
		"CREATE INDEX idx_intNotNullable ON alltypes (intNotNullable) WHERE intNotNullable > 1;\n"+
		"CREATE INDEX idx_textNullable ON alltypes (textNullable);")
}

func (s *DBRootCommandShellSuite) Test_GivenATableWithRecord_WhenCallDotModeCSVAndSelect_ExpectNoErrors() {
	s.tc.CreateSimpleTable("simple_table", []utils.SimpleTableEntry{{TextField: "value \"1", IntField: 1}, {TextField: "value, 2", IntField: 2}})

	outS, errSMode, err := s.tc.ExecuteShell([]string{".mode csv", "SELECT * from simple_table;"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errSMode, qt.Equals, "")
	s.tc.Assert(outS, qt.Equals, "id,textField,intField\n1,\"value \"\"1\",1\n2,\"value, 2\",2")
}

func (s *DBRootCommandShellSuite) Test_GivenAnEmptyTable_WhenCallDotModeCSVAndSelect_ExpectNoErrors() {
	s.tc.CreateEmptySimpleTable("simple_table")

	outS, errSMode, err := s.tc.ExecuteShell([]string{".mode csv", "SELECT * from simple_table;"})
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errSMode, qt.Equals, "")
	s.tc.Assert(outS, qt.Equals, "id,textField,intField")
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
