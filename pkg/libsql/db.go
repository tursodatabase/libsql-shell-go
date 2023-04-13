package libsql

import (
	"database/sql"
	"io"
	"reflect"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xwb1989/sqlparser"

	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
	_ "github.com/libsql/sqld/packages/golang/libsql-client/sql_driver"
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

	return &Db{sqlDb: sqlDb, Path: dbPath}, nil
}

func (db *Db) Close() {
	db.sqlDb.Close()
}

func (db *Db) ExecuteStatements(statementsString string) (StatementsResult, error) {

	statements, err := sqlparser.SplitStatementToPieces(statementsString)
	if err != nil {
		return StatementsResult{}, err
	}

	statementResultCh := make(chan StatementResult)

	go func() {
		defer close(statementResultCh)
		db.executeStatementsAndPopulateChannel(statements, statementResultCh)
	}()

	return StatementsResult{StatementResultCh: statementResultCh}, nil
}

func (db *Db) executeStatementsAndPopulateChannel(statements []string, statementResultCh chan StatementResult) {
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

func (db *Db) executeStatement(statement string, rowsEndedWithoutErrorCh chan bool) *StatementResult {
	if strings.TrimSpace(statement) == "" {
		return nil
	}

	rows, err := db.sqlDb.Query(statement)
	if err != nil {
		return &StatementResult{Err: err}
	}

	columnNames, err := getColumnNames(rows)
	if err != nil {
		return &StatementResult{Err: err}
	}

	columnTypes, err := getColumnTypes(rows)
	if err != nil {
		return &StatementResult{Err: err}
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

	return &StatementResult{ColumnNames: columnNames, RowCh: rowCh}
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
