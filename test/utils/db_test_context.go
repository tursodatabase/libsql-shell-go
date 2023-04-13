package utils

import (
	"fmt"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/libsql/libsql-shell-go/internal/cmd"
)

type DbTestContext struct {
	*testing.T
	*qt.C

	dbPath string
}

func NewTestContext(t *testing.T, dbPath string) *DbTestContext {
	return &DbTestContext{T: t, C: qt.New(t), dbPath: dbPath}
}

func (tc *DbTestContext) Execute(statements string) (string, string, error) {
	rootCmd := cmd.NewRootCmd()
	return Execute(tc.T, rootCmd, "--exec", statements, tc.dbPath)
}

func (tc *DbTestContext) ExecuteShell(commands []string) (outS string, errS string, err error) {
	rootCmd := cmd.NewRootCmd()
	return ExecuteWithInitialInput(tc.T, rootCmd, strings.Join(commands, "\n"), tc.dbPath, "--quiet")
}

func (tc *DbTestContext) CreateEmptySimpleTable(tableName string) {
	_, _, err := tc.Execute("CREATE TABLE " + tableName + " (id INTEGER PRIMARY KEY, textField TEXT, intField INTEGER)")
	tc.Assert(err, qt.IsNil)
}

func (tc *DbTestContext) CreateEmptyAllTypesTable(tableName string) {
	_, _, err := tc.Execute("CREATE TABLE " + tableName + ` (textNullable text, textNotNullable text NOT NULL, textWithDefault text DEFAULT 'defaultValue', 
	intNullable INTEGER, intNotNullable INTEGER NOT NULL, intWithDefault INTEGER DEFAULT '0', 
	floatNullable REAL, floatNotNullable REAL NOT NULL, floatWithDefault REAL DEFAULT '0.0', 
	unknownNullable NUMERIC, unknownNotNullable NUMERIC NOT NULL, unknownWithDefault NUMERIC DEFAULT 0.0, 
	blobNullable BLOB, blobNotNullable BLOB NOT NULL, blobWithDefault BLOB DEFAULT 'x"0"');`)
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

	_, _, err := tc.Execute(insertQuery)
	tc.Assert(err, qt.IsNil)
}

type AllTypesTableEntry struct {
	TextNotNullable    string
	IntNotNullable     int
	FloatNotNullable   float64
	UnknownNotNullable float64
	BlobNotNullable    string
}

func (tc *DbTestContext) CreateAllTypesTable(tableName string, initialValues []AllTypesTableEntry) {
	tc.CreateEmptyAllTypesTable(tableName)

	if len(initialValues) == 0 {
		return
	}

	values := make([]string, 0, len(initialValues))
	for _, initialValue := range initialValues {
		values = append(values,
			fmt.Sprintf("('%v', %v, %f, %f, X'%v')",
				initialValue.TextNotNullable,
				initialValue.IntNotNullable,
				initialValue.FloatNotNullable,
				initialValue.UnknownNotNullable,
				initialValue.BlobNotNullable),
		)
	}
	insertQuery := "INSERT INTO " + tableName + `(textNotNullable, intNotNullable, 
		floatNotNullable, unknownNotNullable, blobNotNullable) VALUES ` + strings.Join(values, ",")

	_, _, err := tc.Execute(insertQuery)
	tc.Assert(err, qt.IsNil)
}

func (tc *DbTestContext) DropTable(tableName string) {
	_, _, err := tc.Execute("DROP TABLE " + tableName)
	tc.Assert(err, qt.IsNil)
}

func (tc *DbTestContext) getAllTables() []string {
	result, _, err := tc.ExecuteShell([]string{".tables"})
	tc.Assert(err, qt.IsNil)
	if strings.TrimSpace(result) == "" {
		return []string{}
	}
	return strings.Split(result, "\n")
}

func (tc *DbTestContext) DropAllTables() {
	for _, createdTable := range tc.getAllTables() {
		tc.DropTable(createdTable)
	}
}

func (tc *DbTestContext) CreateTempFile(content string) (*os.File, string) {
	filePath := tc.C.TempDir() + `/test.txt`
	file, err := os.Create(filePath)
	tc.Assert(err, qt.IsNil)

	_, err = file.WriteString(content)
	tc.Assert(err, qt.IsNil)

	return file, filePath
}
