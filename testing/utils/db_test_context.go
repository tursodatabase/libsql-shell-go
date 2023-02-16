package utils

import (
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/spf13/cobra"

	"github.com/chiselstrike/libsql-shell/src/cmd"
)

type DbTestContext struct {
	*testing.T
	*qt.C

	dbPath  string
	rootCmd *cobra.Command
}

func NewTestContext(t *testing.T) *DbTestContext {
	return &DbTestContext{t, qt.New(t), t.TempDir() + "test.sqlite", cmd.NewRootCmd()}
}

func (tc *DbTestContext) Execute(statements string) (string, error) {
	return Execute(tc.T, tc.rootCmd, "--exec", statements, tc.dbPath)
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
