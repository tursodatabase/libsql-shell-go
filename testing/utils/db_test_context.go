package utils

import (
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/chiselstrike/libsql-shell/src/cmd"
)

type DbTestContext struct {
	*testing.T
	*qt.C

	dbPath string
}

func NewTestContext(t *testing.T, dbPath string) *DbTestContext {
	return &DbTestContext{T: t, C: qt.New(t), dbPath: dbPath}
}

func (tc *DbTestContext) Execute(statements string) (string, error) {
	rootCmd := cmd.NewRootCmd()
	return Execute(tc.T, rootCmd, "--exec", statements, tc.dbPath)
}

func (tc *DbTestContext) ExecuteShell(commands []string) (string, error) {
	rootCmd := cmd.NewRootCmd()
	return ExecuteWithInitialInput(tc.T, rootCmd, strings.Join(commands, "\n"), tc.dbPath, "--quiet")
}

func (tc *DbTestContext) CreateEmptySimpleTable(tableName string) {
	_, err := tc.Execute("CREATE TABLE " + tableName + " (id INTEGER PRIMARY KEY, textField TEXT, intField INTEGER)")
	tc.Assert(err, qt.IsNil)
}

type SimpleTableEntry struct {
	TextField string
	IntField  int
}

func (tc *DbTestContext) CreateSimpleTable(tableName string, initialValues []SimpleTableEntry) {
	tc.CreateEmptySimpleTable(tableName)

	if len(initialValues) == 0 {
		return
	}

	values := make([]string, 0, len(initialValues))
	for _, initialValue := range initialValues {
		values = append(values, fmt.Sprintf("('%v', %v)", initialValue.TextField, initialValue.IntField))
	}
	insertQuery := "INSERT INTO " + tableName + "(textField, intField) VALUES " + strings.Join(values, ",")

	_, err := tc.Execute(insertQuery)
	tc.Assert(err, qt.IsNil)
}

func (tc *DbTestContext) DropTable(tableName string) {
	_, err := tc.Execute("DROP TABLE " + tableName)
	tc.Assert(err, qt.IsNil)
}

func (tc *DbTestContext) getAllTables() []string {
	result, err := tc.ExecuteShell([]string{".tables"})
	tc.Assert(err, qt.IsNil)
	if strings.TrimSpace(result) == "" {
		return []string{}
	}
	return strings.Split(result, "\n")
}

func (tc *DbTestContext) dropAllTables() {
	for _, createdTable := range tc.getAllTables() {
		tc.DropTable(createdTable)
	}
}

func (tc *DbTestContext) TearDown() {
	tc.dropAllTables()
}
