package lib

import (
	"database/sql"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const COLUMN_SEPARATOR = "|"
const STATEMENT_SEPARATOR = ";"

func OpenOrCreateSQLite3(filename string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func ExecuteStatements(db *sql.DB, statementsString string) (string, error) {
	statements := strings.Split(statementsString, STATEMENT_SEPARATOR)
	statementResults := make([]string, 0, len(statements))
	for _, statement := range statements {
		statementResult, err := executeStatement(db, statement)
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

func executeStatement(db *sql.DB, statement string) (string, error) {
	if strings.TrimSpace(statement) == "" {
		return "", nil
	}

	rows, err := db.Query(statement)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	columnNames, err := getColumnNames(rows)
	if err != nil {
		return "", err
	}

	result := strings.Join(columnNames, COLUMN_SEPARATOR)

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

		result = result + "\n" + strings.Join(columnValues, COLUMN_SEPARATOR)
	}

	return result, nil
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
