package lib

import (
	"database/sql"
	"io"
	"reflect"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xwb1989/sqlparser"

	_ "github.com/libsql/sqld/packages/golang/libsql-client/sql_driver"
)

type Db struct {
	sqlDb *sql.DB
	path  string
}

type statementsResult struct {
	StatementResultCh chan statementResult
}

type statementResult struct {
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
	if IsHttpUrl(dbPath) {
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

	return &Db{sqlDb: sqlDb, path: dbPath}, nil
}

func (db *Db) Close() {
	db.sqlDb.Close()
}

func (db *Db) ExecuteStatements(statementsString string) (statementsResult, error) {

	statements, err := sqlparser.SplitStatementToPieces(statementsString)
	if err != nil {
		return statementsResult{}, err
	}

	statementResultCh := make(chan statementResult)

	go func() {
		defer close(statementResultCh)
		db.executeStatementsAndPopulateChannel(statements, statementResultCh)
	}()

	return statementsResult{StatementResultCh: statementResultCh}, nil
}

func (db *Db) executeStatementsAndPopulateChannel(statements []string, statementResultCh chan statementResult) {
	rowsEndedWithoutErrorCh := make(chan bool)
	defer close(rowsEndedWithoutErrorCh)

	for _, statement := range statements {
		result := db.executeStatement(statement, rowsEndedWithoutErrorCh)

		if result == nil {
			continue
		}

		if result.Err != nil {
			if strings.Contains(result.Err.Error(), "interactive transaction not allowed in HTTP queries") {
				result.Err = &TransactionNotSupportedError{}
			}
			statementResultCh <- *result
			return
		}

		statementResultCh <- *result
		shouldContinue := <-rowsEndedWithoutErrorCh
		if !shouldContinue {
			return
		}
	}
}

func (db *Db) ExecuteAndPrintStatements(statementsString string, outF io.Writer, errF io.Writer, withoutHeader bool) {
	result, err := db.ExecuteStatements(statementsString)
	if err != nil {
		PrintError(err, errF)
		return
	}

	err = PrintStatementsResult(result, outF, withoutHeader)
	if err != nil {
		PrintError(err, errF)
		return
	}
}

func (db *Db) executeStatement(statement string, rowsEndedWithoutErrorCh chan bool) *statementResult {
	if strings.TrimSpace(statement) == "" {
		return nil
	}

	rows, err := db.sqlDb.Query(statement)
	if err != nil {
		return &statementResult{Err: err}
	}

	columnNames, err := getColumnNames(rows)
	if err != nil {
		return &statementResult{Err: err}
	}

	columnTypes, err := getColumnTypes(rows)
	if err != nil {
		return &statementResult{Err: err}
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
		defer rows.Close()
		defer close(rowCh)

		for rows.Next() {
			err = rows.Scan(columnPointers...)
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

		if err := rows.Err(); err != nil {
			rowCh <- rowResult{Err: err}
			rowsEndedWithoutErrorCh <- false
			return
		}

		rowsEndedWithoutErrorCh <- true
	}()

	return &statementResult{ColumnNames: columnNames, RowCh: rowCh}
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
