package lib

import (
	"database/sql"
	"fmt"
	"io"
	"net/url"
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
	Data        [][]string
}

const COLUMN_SEPARATOR = "|"

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

	PrintStatementsResults(results, outF, withoutHeader)
}

func PrintStatementsResults(results []Result, outF io.Writer, withoutHeader bool) {
	for _, result := range results {
		if len(result.ColumnNames) != 0 {
			if withoutHeader {
				PrintTable(outF, nil, result.Data)
			} else {
				PrintTable(outF, result.ColumnNames, result.Data)
			}
		}
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

	data := make([][]string, 0)

	columnNamesLen := len(columnNames)
	columnValues := make([]string, columnNamesLen)
	columnPointers := make([]interface{}, columnNamesLen)
	for i := range columnNames {
		columnPointers[i] = &columnValues[i]
	}

	for rows.Next() {
		err = rows.Scan(columnPointers...)
		if err != nil {
			return nil, err
		}

		currentRow := make([]string, columnNamesLen)
		copy(currentRow, columnValues)
		data = append(data, currentRow)
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
