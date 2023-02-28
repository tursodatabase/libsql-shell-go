package main_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/stretchr/testify/suite"

	"github.com/chiselstrike/libsql-shell/testing/utils"
)

type RootCommandExecSuite struct {
	suite.Suite

	// dbInTest utils.DbType
	dbPath string
	tc     *utils.DbTestContext
}

func NewRootCommandExecSuite(dbPath string) *RootCommandExecSuite {
	return &RootCommandExecSuite{dbPath: dbPath}
}

func (s *RootCommandExecSuite) SetupTest() {
	s.tc = utils.NewTestContext(s.T(), s.dbPath)
}

func (s *RootCommandExecSuite) TearDownTest() {
	s.tc.TearDown()
}

func (s *RootCommandExecSuite) TestRootCommandExec_GivenEmptyDb_WhenCreateTable_ExpectEmptyResult() {
	result, err := s.tc.Execute("CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT)")

	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(result, qt.Equals, "")
}

func (s *RootCommandExecSuite) TestRootCommandExec_GivenSimpleTableCreated_WhenInsertValue_ExpectEmptyResult() {
	s.tc.CreateEmptySimpleTable("simple_table")

	result, err := s.tc.Execute("INSERT INTO simple_table(textField, intField) VALUES ('textValue', 1)")

	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(result, qt.Equals, "")
}

func (s *RootCommandExecSuite) TestRootCommandExec_GivenSimpleTableCreated_WhenSelectEntireTable_ExpectFirstLineToBeTheHeader() {
	s.tc.CreateEmptySimpleTable("simple_table")

	result, err := s.tc.Execute("SELECT * FROM simple_table")
	s.tc.Assert(err, qt.IsNil)

	headerLine := strings.Split(result, "\n")[0]

	s.tc.Assert(headerLine, qt.Equals, utils.GetPrintTableOutput([]string{"id", "textField", "intField"}, [][]interface{}{}))
}

func (s *RootCommandExecSuite) TestRootCommandExec_GivenPopulatedSimpleTable_WhenSelectEntireTable_ExpectAllEntries() {
	s.tc.CreateSimpleTable("simple_table", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})

	result, err := s.tc.Execute("SELECT * FROM simple_table")
	s.tc.Assert(err, qt.IsNil)

	s.tc.Assert(result, qt.Equals, utils.GetPrintTableOutput([]string{"id", "textField", "intField"}, [][]interface{}{{"1", "value1", "1"}, {"2", "value2", "2"}}))
}

func (s *RootCommandExecSuite) TestRootCommandExec_GivenPopulatedSimpleTable_WhenSelectEntireTableTwice_ExpectTwoResults() {
	s.tc.CreateSimpleTable("simple_table", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})

	result, err := s.tc.Execute("SELECT * FROM simple_table; SELECT * FROM simple_table")
	s.tc.Assert(err, qt.IsNil)

	resultText := utils.GetPrintTableOutput([]string{"id", "textField", "intField"}, [][]interface{}{{"1", "value1", "1"}, {"2", "value2", "2"}})
	resultLines := resultText + "            \n" + resultText
	s.tc.Assert(result, qt.ContentEquals, resultLines)
}

func (s *RootCommandExecSuite) TestRootCommandExec_GivenEmptyDb_WhenCreateInsertAndSelectTableAtOnce_ExpectSelectResult() {

	result, err := s.tc.Execute("CREATE TABLE simple_table (id INTEGER PRIMARY KEY, textField TEXT, intField INTEGER); INSERT INTO simple_table(textField, intField) VALUES ('value1', 1), ('value2', 2); SELECT * FROM simple_table")
	s.tc.Assert(err, qt.IsNil)

	s.tc.Assert(result, qt.Equals, utils.GetPrintTableOutput([]string{"id", "textField", "intField"}, [][]interface{}{{"1", "value1", "1"}, {"2", "value2", "2"}}))
}

func (s *RootCommandExecSuite) TestRootCommandExec_WhenSendStatementWithSemicolonAtEnd_ExpectNoError() {
	s.tc.CreateSimpleTable("simple_table", []utils.SimpleTableEntry{{TextField: "value1", IntField: 1}, {TextField: "value2", IntField: 2}})

	result, err := s.tc.Execute("SELECT * FROM simple_table;;;;;;;")
	s.tc.Assert(err, qt.IsNil)

	s.tc.Assert(result, qt.Equals, utils.GetPrintTableOutput([]string{"id", "textField", "intField"}, [][]interface{}{{"1", "value1", "1"}, {"2", "value2", "2"}}))

	s.tc.Assert(err, qt.IsNil)
}

func (s *RootCommandExecSuite) TestRootCommandExec_GivenSimpleTableCreated_WhenInsertValueWithSemiColumnAndSelectIt_ExpectNoError() {
	s.tc.CreateEmptySimpleTable("simple_table")

	result, err := s.tc.Execute("INSERT INTO simple_table(textField, intField) VALUES ('text;Value', 1)")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(result, qt.Equals, "")

	result, err = s.tc.Execute("SELECT * FROM simple_table")
	s.tc.Assert(err, qt.IsNil)

	s.tc.Assert(result, qt.Equals, utils.GetPrintTableOutput([]string{"id", "textField", "intField"}, [][]interface{}{{"1", "text;Value", "1"}}))
}

func (s *RootCommandExecSuite) TestRootCommandExec_GivenEmptyDB_WhenSelectNull_ExpectNULLAsReturn() {
	result, err := s.tc.Execute("SELECT NULL")
	s.tc.Assert(err, qt.IsNil)

	s.tc.Assert(result, qt.Equals, utils.GetPrintTableOutput([]string{"NULL"}, [][]interface{}{{"NULL"}}))
}

func (s *RootCommandExecSuite) TestRootCommandExec_GivenTableCotainingBlobField_WhenInsertAndSelect_ExpectNoError() {
	_, err := s.tc.Execute("CREATE TABLE alltypes (t text, i integer, r real, b blob)")
	s.tc.Assert(err, qt.IsNil)

	_, err = s.tc.Execute("INSERT INTO alltypes VALUES ('text', 99, 3.14, x'0123456789ABCDEF')")
	s.tc.Assert(err, qt.IsNil)

	result, err := s.tc.Execute("SELECT * from alltypes")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(result, qt.Equals, utils.GetPrintTableOutput([]string{"T", "I", "R", "B"}, [][]interface{}{{"text", "99", "3.14", "0123456789abcdef"}}))
}

func TestRootCommandExecSuite_WhenDbIsSQLite(t *testing.T) {
	suite.Run(t, NewRootCommandExecSuite(t.TempDir()+"test.sqlite"))
}
