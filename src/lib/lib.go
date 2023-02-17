package lib

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xwb1989/sqlparser"
)

type dbOptions struct {
	withoutHeader bool
}
type Db struct {
	sqlDb   *sql.DB
	options dbOptions
}

const COLUMN_SEPARATOR = "|"

func NewSQLite3(filename string) (*Db, error) {
	sqlDb, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	if err = sqlDb.Ping(); err != nil {
		return nil, err
	}

	return &Db{sqlDb: sqlDb}, nil
}

func (db *Db) Close() {
	db.sqlDb.Close()
}

func (db *Db) ExecuteStatements(statementsString string) (string, error) {
	statements, err := sqlparser.SplitStatementToPieces(statementsString)
	if err != nil {
		return "", err
	}

	statementResults := make([]string, 0, len(statements))
	for _, statement := range statements {
		statementResult, err := db.executeStatement(statement)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(statementResult) != "" {
			statementResults = append(statementResults, statementResult)
		}
	}

	allStatementResults := strings.Join(statementResults, "\n")
	return allStatementResults, nil
}

func (db *Db) ExecuteAndPrintStatements(statementsString string) {
	result, err := db.ExecuteStatements(statementsString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	} else {
		fmt.Println(result)
	}
}

func (db *Db) executeStatement(statement string) (string, error) {
	if strings.TrimSpace(statement) == "" {
		return "", nil
	}

	rows, err := db.sqlDb.Query(statement)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	columnNames, err := getColumnNames(rows)
	if err != nil {
		return "", err
	}

	result := make([]string, 0)

	if !db.options.withoutHeader {
		result = append(result, strings.Join(columnNames, COLUMN_SEPARATOR))
	}

	columnValues := make([]string, len(columnNames))
	columnPointers := make([]interface{}, len(columnNames))
	for i := range columnNames {
		columnPointers[i] = &columnValues[i]
	}

	for rows.Next() {
		err = rows.Scan(columnPointers...)
		if err != nil {
			return "", err
		}

		result = append(result, strings.Join(columnValues, COLUMN_SEPARATOR))
	}

	return strings.Join(result, "\n"), nil
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
