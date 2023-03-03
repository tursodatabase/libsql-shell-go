package lib

import (
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xwb1989/sqlparser"

	_ "github.com/libsql/sqld/packages/golang/libsql-client/sql_driver"
)

type Db struct {
	sqlDb *sql.DB
}
type Result struct {
	ColumnNames []string
	Data        [][]interface{}
}

func NewDb(dbPath string) (*Db, error) {
	var sqlDb *sql.DB
	var err error
	if isHttpUrl(dbPath) {
		sqlDb, err = sql.Open("libsql", dbPath)
	} else {
		sqlDb, err = sql.Open("sqlite3", dbPath)
	}
	if err != nil {
		return nil, err
	}

	if err = sqlDb.Ping(); err != nil {
		return nil, err
	}

	return &Db{sqlDb: sqlDb}, nil
}

func isHttpUrl(path string) bool {
	url, err := url.ParseRequestURI(path)
	if err != nil {
		return false
	}
	return url.Scheme == "http" || url.Scheme == "https"
}

func (db *Db) Close() {
	db.sqlDb.Close()
}

func (db *Db) ExecuteStatements(statementsString string) ([]Result, error) {
	statements, err := sqlparser.SplitStatementToPieces(statementsString)
	if err != nil {
		return nil, err
	}

	statementResults := make([]Result, 0, len(statements))
	for _, statement := range statements {
		result, err := db.executeStatement(statement)
		if err != nil {
			if strings.Contains(err.Error(), "interactive transaction not allowed in HTTP queries") {
				return nil, &TransactionNotSupportedError{}
			}
			return nil, err
		}
		if result != nil {
			statementResults = append(statementResults, *result)
		}
	}

	return statementResults, nil
}

func (db *Db) ExecuteAndPrintStatements(statementsString string, outF io.Writer, errF io.Writer, withoutHeader bool) {
	results, err := db.ExecuteStatements(statementsString)
	if err != nil {
		fmt.Fprintf(errF, "Error: %s\n", err.Error())
		return
	}

	err = PrintStatementsResults(results, outF, withoutHeader)
	if err != nil {
		fmt.Fprintf(errF, "Error: %s\n", err.Error())
		return
	}
}

func (db *Db) executeStatement(statement string) (*Result, error) {
	if strings.TrimSpace(statement) == "" {
		return nil, nil
	}

	rows, err := db.sqlDb.Query(statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columnNames, err := getColumnNames(rows)
	if err != nil {
		return nil, err
	}

	columnTypes, err := getColumnTypes(rows)
	if err != nil {
		return nil, err
	}

	data := make([][]interface{}, 0)

	columnNamesLen := len(columnNames)

	columnPointers := make([]interface{}, columnNamesLen)
	for i, t := range columnTypes {
		if t.Kind() == reflect.Struct {
			columnPointers[i] = reflect.New(t).Interface()
		} else {
			columnPointers[i] = new(interface{})
		}
	}

	for rows.Next() {
		err = rows.Scan(columnPointers...)
		if err != nil {
			return nil, err
		}

		rowData := make([]interface{}, len(columnTypes))
		for i, ptr := range columnPointers {
			val := reflect.ValueOf(ptr).Elem()
			rowData[i] = val.Interface()
		}
		data = append(data, rowData)
	}

	return &Result{ColumnNames: columnNames, Data: data}, nil
}

func getColumnNames(rows *sql.Rows) ([]string, error) {
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	columnNames := make([]string, 0, len(columnTypes))
	for _, columnType := range columnTypes {
		columnNames = append(columnNames, columnType.Name())
	}

	return columnNames, nil
}

func getColumnTypes(rows *sql.Rows) ([]reflect.Type, error) {
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	types := make([]reflect.Type, len(columnTypes))

	for i, ct := range columnTypes {
		types[i] = ct.ScanType()
	}

	return types, nil
}
