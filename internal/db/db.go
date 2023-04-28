package db

import (
	"database/sql"
	"io"
	"reflect"
	"strings"

	_ "github.com/libsql/libsql-client-go/libsql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xwb1989/sqlparser"

	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
	"github.com/libsql/libsql-shell-go/pkg/shell/shellerrors"
)

type Db struct {
	sqlDb *sql.DB
	Path  string
}

type StatementsResult struct {
	StatementResultCh chan StatementResult
}

type StatementResult struct {
	ColumnNames []string
	RowCh       chan rowResult
	Err         error
}

type rowResult struct {
	Row []interface{}
	Err error
}

func NewDb(dbPath string) (*Db, error) {
	var sqlDb *sql.DB
	var err error
	if IsUrl(dbPath) {
		if IsValidTursoUrl(dbPath) {
			sqlDb, err = sql.Open("libsql", dbPath)
		} else {
			return nil, &shellerrors.InvalidTursoProtocolError{}
		}
	} else {
		sqlDb, err = sql.Open("sqlite3", dbPath)
	}
	if err != nil {
		return nil, err
	}

	if err = sqlDb.Ping(); err != nil {
		return nil, err
	}

	return &Db{sqlDb: sqlDb, Path: dbPath}, nil
}

func (db *Db) Close() {
	db.sqlDb.Close()
}

func (db *Db) ExecuteStatements(statementsString string) (StatementsResult, error) {

	queries, err := sqlparser.SplitStatementToPieces(statementsString)
	if err != nil {
		return StatementsResult{}, err
	}

	statementResultCh := make(chan StatementResult)

	go func() {
		defer close(statementResultCh)
		db.executeQueriesAndPopulateChannel(queries, statementResultCh)
	}()

	return StatementsResult{StatementResultCh: statementResultCh}, nil
}

func (db *Db) executeQueriesAndPopulateChannel(queries []string, statementResultCh chan StatementResult) {
	for _, query := range queries {
		if shouldContinue := db.executeQuery(query, statementResultCh); !shouldContinue {
			return
		}
	}
}

func (db *Db) ExecuteAndPrintStatements(statementsString string, outF io.Writer, withoutHeader bool, printMode enums.PrintMode) error {
	result, err := db.ExecuteStatements(statementsString)
	if err != nil {
		return err
	}

	err = PrintStatementsResult(result, outF, withoutHeader, printMode)
	if err != nil {
		return err
	}

	return nil
}

func (db *Db) executeQuery(query string, statementResultCh chan StatementResult) (queryEndedWithoutError bool) {
	if strings.TrimSpace(query) == "" {
		return true
	}

	rows, err := db.sqlDb.Query(query)
	if err != nil {
		if strings.Contains(err.Error(), "interactive transaction not allowed in HTTP queries") {
			err = &shellerrors.TransactionNotSupportedError{}
		}

		statementResultCh <- StatementResult{Err: err}
		return false
	}
	defer rows.Close()

	return readQueryResults(rows, statementResultCh)
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

func readQueryResults(queryRows *sql.Rows, statementResultCh chan StatementResult) (queryEndedWithoutError bool) {
	rowsEndedWithoutErrorCh := make(chan bool)
	defer close(rowsEndedWithoutErrorCh)

	hasResultSetToRead := true
	for hasResultSetToRead {
		statementResult := readQueryResultSet(queryRows, rowsEndedWithoutErrorCh)
		statementResultCh <- statementResult

		if statementResult.Err != nil {
			return false
		}

		shouldContinue := <-rowsEndedWithoutErrorCh
		if !shouldContinue {
			return false
		}

		hasResultSetToRead = queryRows.NextResultSet()
	}

	if err := queryRows.Err(); err != nil {
		statementResultCh <- StatementResult{Err: err}
		return false
	}

	return true
}

func readQueryResultSet(queryRows *sql.Rows, rowsEndedWithoutErrorCh chan bool) StatementResult {
	columnNames, err := getColumnNames(queryRows)
	if err != nil {
		return StatementResult{Err: err}
	}

	columnTypes, err := getColumnTypes(queryRows)
	if err != nil {
		return StatementResult{Err: err}
	}

	columnNamesLen := len(columnNames)
	columnPointers := make([]interface{}, columnNamesLen)
	for i, t := range columnTypes {
		if t.Kind() == reflect.Struct {
			columnPointers[i] = reflect.New(t).Interface()
		} else {
			columnPointers[i] = new(interface{})
		}
	}

	rowCh := make(chan rowResult)
	go func() {
		defer close(rowCh)

		for queryRows.Next() {
			err = queryRows.Scan(columnPointers...)
			if err != nil {
				rowCh <- rowResult{Err: err}
				rowsEndedWithoutErrorCh <- false
				return
			}

			rowData := make([]interface{}, len(columnTypes))
			for i, ptr := range columnPointers {
				val := reflect.ValueOf(ptr).Elem()
				rowData[i] = val.Interface()
			}
			rowCh <- rowResult{Row: rowData}
		}

		if err := queryRows.Err(); err != nil {
			rowCh <- rowResult{Err: err}
			rowsEndedWithoutErrorCh <- false
			return
		}

		rowsEndedWithoutErrorCh <- true
	}()

	return StatementResult{ColumnNames: columnNames, RowCh: rowCh}
}
