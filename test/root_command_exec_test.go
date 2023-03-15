package main_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/stretchr/testify/suite"

	"github.com/chiselstrike/libsql-shell/test/utils"
)

type RootCommandExecSuite struct {
	suite.Suite

	dbPath string
	tc     *utils.DbTestContext
}

func NewRootCommandExecSuite(dbPath string) *RootCommandExecSuite {
	return &RootCommandExecSuite{dbPath: dbPath}
}

func (s *RootCommandExecSuite) SetupSuite() {
	s.tc = utils.NewTestContext(s.T(), s.dbPath)
	s.tc.DropAllTables()
}

func (s *RootCommandExecSuite) TearDownTest() {
	s.tc.DropAllTables()
}

func (s *RootCommandExecSuite) Test_GivenEmptyDb_WhenCreateTable_ExpectEmptyResult() {
	outS, errS, err := s.tc.Execute("CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT)")

	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(outS, qt.Equals, "")
	s.tc.Assert(errS, qt.Equals, "")
}

func (s *RootCommandExecSuite) Test_GivenSimpleTableCreated_WhenInsertValue_ExpectEmptyResult() {
	s.tc.CreateEmptySimpleTable("simple_table")

	outS, errS, err := s.tc.Execute("INSERT INTO simple_table(textField, intField) VALUES ('textValue', 1)")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, "")
}

func (s *RootCommandExecSuite) Test_GivenSimpleTableCreated_WhenSelectEntireTable_ExpectFirstLineToBeTheHeader() {
	s.tc.CreateEmptySimpleTable("simple_table")

	outS, errS, err := s.tc.Execute("SELECT * FROM simple_table")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	headerLine := strings.Split(outS, "\n")[0]

	s.tc.Assert(headerLine, qt.Equals, utils.GetPrintTableOutput([]string{"id", "textField", "intField"}, [][]string{}))
}

func (s *RootCommandExecSuite) Test_GivenPopulatedSimpleTable_WhenSelectEntireTable_ExpectAllEntries() {
	s.tc.CreateSimpleTable("simple_table", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})

	outS, errS, err := s.tc.Execute("SELECT * FROM simple_table")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{"id", "textField", "intField"}, [][]string{{"1", "value1", "1"}, {"2", "value2", "2"}}))
}

func (s *RootCommandExecSuite) Test_GivenPopulatedSimpleTable_WhenSelectEntireTableTwice_ExpectTwoResults() {
	s.tc.CreateSimpleTable("simple_table", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})

	outS, errS, err := s.tc.Execute("SELECT * FROM simple_table; SELECT * FROM simple_table")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	resultText := utils.GetPrintTableOutput([]string{"id", "textField", "intField"}, [][]string{{"1", "value1", "1"}, {"2", "value2", "2"}})
	resultLines := resultText + "            \n" + resultText
	s.tc.Assert(outS, qt.ContentEquals, resultLines)
}

func (s *RootCommandExecSuite) Test_GivenEmptyDb_WhenCreateInsertAndSelectTableAtOnce_ExpectSelectResult() {

	outS, errS, err := s.tc.Execute("CREATE TABLE simple_table (id INTEGER PRIMARY KEY, textField TEXT, intField INTEGER); INSERT INTO simple_table(textField, intField) VALUES ('value1', 1), ('value2', 2); SELECT * FROM simple_table")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{"id", "textField", "intField"}, [][]string{{"1", "value1", "1"}, {"2", "value2", "2"}}))
}

func (s *RootCommandExecSuite) Test_WhenSendStatementWithSemicolonAtEnd_ExpectNoError() {
	s.tc.CreateSimpleTable("simple_table", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})

	outS, errS, err := s.tc.Execute("SELECT * FROM simple_table;;;;;;;")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{"id", "textField", "intField"}, [][]string{{"1", "value1", "1"}, {"2", "value2", "2"}}))

	s.tc.Assert(err, qt.IsNil)
}

func (s *RootCommandExecSuite) Test_GivenSimpleTableCreated_WhenInsertValueWithSemiColumnAndSelectIt_ExpectNoError() {
	s.tc.CreateEmptySimpleTable("simple_table")

	outS, errS, err := s.tc.Execute("INSERT INTO simple_table(textField, intField) VALUES ('text;Value', 1)")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outS, qt.Equals, "")

	outS, errS, err = s.tc.Execute("SELECT * FROM simple_table")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{"id", "textField", "intField"}, [][]string{{"1", "text;Value", "1"}}))
}

func (s *RootCommandExecSuite) TestRootCommandExec_GivenEmptyDB_WhenSelectNull_ExpectNULLAsReturn() {
	outS, errS, err := s.tc.Execute("SELECT NULL")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{"NULL"}, [][]string{{"NULL"}}))
}

func (s *RootCommandExecSuite) TestRootCommandExec_GivenTableCotainingBlobField_WhenInsertAndSelect_ExpectNoError() {
	_, errS, err := s.tc.Execute("CREATE TABLE alltypes (t text, i integer, r real, b blob)")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	_, errS, err = s.tc.Execute("INSERT INTO alltypes VALUES ('text', 99, 3.14, x'0123456789ABCDEF')")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	outS, errS, err := s.tc.Execute("SELECT * from alltypes")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{"T", "I", "R", "B"}, [][]string{{"text", "99", "3.14", "0x0123456789ABCDEF"}}))
}

func TestRootCommandExecSuite_WhenDbIsSQLite(t *testing.T) {
	suite.Run(t, NewRootCommandExecSuite(t.TempDir()+"test.sqlite"))
}

func TestRootCommandExecSuite_WhenDbIsTurso(t *testing.T) {
	testConfig := utils.GetTestConfig(t)
	if testConfig.SkipTursoTests {
		t.Skip("Skipping Turso tests due configuration")
	}

	suite.Run(t, NewRootCommandExecSuite(testConfig.TursoDbPath))
}
