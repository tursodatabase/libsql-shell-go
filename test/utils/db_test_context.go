package utils

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/libsql/libsql-shell-go/internal/db"
	"github.com/libsql/libsql-shell-go/internal/shell"
	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
)

type DbTestContext struct {
	*testing.T
	*qt.C

	dbUri string

	db *db.Db
}

func NewTestContext(t *testing.T, dbUri string, authToken string) *DbTestContext {
	db, err := db.NewDb(dbUri, authToken, "", false, "")
	if err != nil {
		t.Fatalf("Fail to create new db. err: %v", err)
	}
	if err := db.TestConnection(); err != nil {
		t.Fatalf("Fail to test connection. err: %v", err)
	}

	return &DbTestContext{T: t, C: qt.New(t), dbUri: dbUri, db: db}
}

func (tc *DbTestContext) Close() {
	tc.db.Close()
}

func (tc *DbTestContext) Execute(statements string) (string, string, error) {
	bufOut, bufErr, config := tc.createShellConfig("")
	shellInstance, err := shell.NewShell(config, tc.db)
	if err != nil {
		tc.T.Fatalf("Fail to create new shell")
	}

	executionError := shellInstance.ExecuteCommandOrStatements(statements)
	return strings.TrimSpace(bufOut.String()), strings.TrimSpace(bufErr.String()), executionError
}

func (tc *DbTestContext) ExecuteShell(commands []string) (outS string, errS string, err error) {
	bufOut, bufErr, config := tc.createShellConfig(strings.Join(commands, "\n"))
	shellInstance, err := shell.NewShell(config, tc.db)
	if err != nil {
		tc.T.Fatalf("Fail to create new shell")
	}

	executionError := shellInstance.Run()
	return strings.TrimSpace(bufOut.String()), strings.TrimSpace(bufErr.String()), executionError
}

func (tc *DbTestContext) createShellConfig(initialInput string) (bufOut *bytes.Buffer, bufErr *bytes.Buffer, config shell.ShellConfig) {
	bufOut = new(bytes.Buffer)
	bufErr = new(bytes.Buffer)
	bufIn := new(bytes.Buffer)

	_, err := bufIn.Write([]byte(initialInput))
	if err != nil {
		tc.T.Fatalf("Fail to write inside initial buffer")
	}

	config = shell.ShellConfig{InF: bufIn, OutF: bufOut, ErrF: bufErr, HistoryMode: enums.SingleHistory, QuietMode: true}
	return
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
	name := tableName
	if db.NeedsEscaping(tableName) {
		name = fmt.Sprintf("'%s'", db.EscapeSingleQuotes(tableName))
	}
	_, _, err := tc.Execute("DROP TABLE " + name)
	tc.Assert(err, qt.IsNil)
}

func (tc *DbTestContext) getAllTables() []string {
	result, _, err := tc.ExecuteShell([]string{".tables"})
	tc.Assert(err, qt.IsNil)
	if strings.TrimSpace(result) == "" {
		return []string{}
	}
	names := strings.Split(result, "\n")
	for i, name := range names {
		names[i] = strings.TrimSpace(name)
	}
	return names
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

func (tc *DbTestContext) AssertSqlEquals(got, want string) bool {
	return tc.Assert(strings.ToLower(got), qt.Equals, strings.ToLower(want))
}
