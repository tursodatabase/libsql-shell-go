package main_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/stretchr/testify/suite"

	"github.com/libsql/libsql-shell-go/test/utils"
)

type RootCommandExecSuite struct {
	suite.Suite

	dbUri     string
	authToken string
	tc        *utils.DbTestContext
}

func NewRootCommandExecSuite(dbUri string) *RootCommandExecSuite {
	return &RootCommandExecSuite{dbUri: dbUri}
}

func (s *RootCommandExecSuite) SetupSuite() {
	s.tc = utils.NewTestContext(s.T(), s.dbUri, s.authToken)
	s.tc.DropAllTables()
}

func (s *RootCommandExecSuite) TearDownSuite() {
	s.tc.Close()
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

func (s *RootCommandExecSuite) Test_GivenEmptyDB_WhenSelectNull_ExpectNULLAsReturn() {
	outS, errS, err := s.tc.Execute("SELECT NULL")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{"NULL"}, [][]string{{"NULL"}}))
}

func (s *RootCommandExecSuite) Test_GivenTableCotainingBlobField_WhenInsertAndSelect_ExpectNoError() {
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

func (s *RootCommandExecSuite) Test_GivenTriggerOnUpdatedAtField_WhenRowUpdated_ExpectUpdatedAtFieldUpdated() {
	_, errS, err := s.tc.Execute("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, updated_at INTEGER)")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	_, errS, err = s.tc.Execute("CREATE TRIGGER update_updated_at AFTER UPDATE ON users FOR EACH ROW BEGIN UPDATE users SET updated_at = 0 WHERE id = NEW.id; END")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	_, errS, err = s.tc.Execute("INSERT INTO users (name) VALUES ('user1')")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	outS, errS, err := s.tc.Execute("SELECT * FROM users")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{"ID", "NAME", "UPDATED AT"}, [][]string{{"1", "user1", "NULL"}}))

	_, errS, err = s.tc.Execute("UPDATE users SET name = 'new_user1' where id=1")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")

	outS, errS, err = s.tc.Execute("SELECT * FROM users")
	s.tc.Assert(err, qt.IsNil)
	s.tc.Assert(errS, qt.Equals, "")
	s.tc.Assert(outS, qt.Equals, utils.GetPrintTableOutput([]string{"ID", "NAME", "UPDATED AT"}, [][]string{{"1", "new_user1", "0"}}))
}

func TestRootCommandExecSuite_WhenDbIsSQLite(t *testing.T) {
	suite.Run(t, NewRootCommandExecSuite(t.TempDir()+"test.sqlite"))
}

func TestRootCommandExecSuite_WhenDbIsSqld(t *testing.T) {
	testConfig := utils.GetTestConfig(t)
	if testConfig.SkipSqldTests {
		t.Skip("Skipping SQLD tests due configuration")
	}

	suite.Run(t, NewRootCommandExecSuite(testConfig.SqldDbUri))
}
