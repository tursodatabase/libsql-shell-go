package db

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"

	"github.com/libsql/sqlite-antlr4-parser/sqliteparserutils"
	_ "github.com/tursodatabase/go-libsql"

	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
	"github.com/libsql/libsql-shell-go/pkg/shell/shellerrors"
)

type Db struct {
	Uri       string
	AuthToken string

	sqlDb    *sql.DB
	isRemote bool

	cancelRunningQuery func()
}

func (db *Db) IsRemote() bool {
	return db.isRemote
}

type StatementsResult struct {
	StatementResultCh chan StatementResult
}

type StatementResult struct {
	ColumnNames []string
	RowCh       chan rowResult
	Err         error
}

func newStatementResult(columnNames []string, rowCh chan rowResult) *StatementResult {
	return &StatementResult{ColumnNames: columnNames, RowCh: rowCh}
}

func newStatementResultWithError(err error) *StatementResult {
	treatedErr := treatDbError(err)
	return &StatementResult{Err: treatedErr}
}

type rowResult struct {
	Row []interface{}
	Err error
}

func newRowResult(row []interface{}) *rowResult {
	return &rowResult{Row: row}
}

func newRowResultWithError(err error) *rowResult {
	treatedErr := treatDbError(err)
	return &rowResult{Err: treatedErr}
}

func NewDb(dbUri, authToken, _proxy string, _schemaDb bool, remoteEncryptionKey string) (*Db, error) {
	var db = Db{Uri: dbUri, AuthToken: authToken}

	// Determine the type of database connection
	connStr := dbUri
	if IsUrl(dbUri) {
		isRemote, scheme := IsValidSqldUrl(dbUri)
		if isRemote {
			// Remote URL (libsql://, http://, https://)
			db.isRemote = true
			if authToken != "" || remoteEncryptionKey != "" {
				u, err := url.Parse(dbUri)
				if err != nil {
					return nil, err
				}
				q := u.Query()
				if authToken != "" {
					q.Set("authToken", authToken)
				}
				if remoteEncryptionKey != "" {
					q.Set("remoteEncryptionKey", remoteEncryptionKey)
				}
				u.RawQuery = q.Encode()
				connStr = u.String()
			}
		} else if scheme == "file" {
			// Local file URL - use as-is
			db.isRemote = false
		} else {
			return nil, &shellerrors.ProtocolError{}
		}
	} else {
		// Plain path - treat as local file, convert to file: URL for go-libsql
		db.isRemote = false
		connStr = "file:" + dbUri
	}

	var err error
	db.sqlDb, err = sql.Open("libsql", connStr)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func (db *Db) TestConnection() error {
	row := db.sqlDb.QueryRow("SELECT 1")
	var result int
	err := row.Scan(&result)
	if err != nil {
		return fmt.Errorf("failed to connect to database. err: %v", err)
	}
	return nil
}

func (db *Db) Close() {
	db.sqlDb.Close()
}

func (db *Db) ExecuteStatements(statementsString string) (StatementsResult, error) {
	queries := db.prepareStatementsIntoQueries(statementsString)

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

	ctx, cancel := context.WithCancel(context.Background())
	db.cancelRunningQuery = cancel

	rows, err := db.sqlDb.QueryContext(ctx, query)
	if err != nil {
		statementResultCh <- *newStatementResultWithError(err)

		return false
	}

	defer rows.Close()

	return readQueryResults(rows, statementResultCh)
}

func (db *Db) prepareStatementsIntoQueries(statementsString string) []string {
	// TODO: check if this required
	stmts, _ := sqliteparserutils.SplitStatement(statementsString)
	return stmts
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

func readQueryResults(queryRows *sql.Rows, statementResultCh chan StatementResult) (shouldContinue bool) {
	hasResultSetToRead := true
	for hasResultSetToRead {
		if shouldContinue := readQueryResultSet(queryRows, statementResultCh); !shouldContinue {
			return false
		}

		hasResultSetToRead = queryRows.NextResultSet()
	}

	if err := queryRows.Err(); err != nil {
		statementResultCh <- *newStatementResultWithError(err)
		return false
	}

	return true
}

func readQueryResultSet(queryRows *sql.Rows, statementResultCh chan StatementResult) (shouldContinue bool) {
	columnNames, err := getColumnNames(queryRows)
	if err != nil {
		statementResultCh <- *newStatementResultWithError(err)
		return false
	}

	columnTypes, err := getColumnTypes(queryRows)
	if err != nil {
		statementResultCh <- *newStatementResultWithError(err)
		return false
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
	defer close(rowCh)

	statementResultCh <- *newStatementResult(columnNames, rowCh)

	for queryRows.Next() {
		err = queryRows.Scan(columnPointers...)
		if err != nil {
			rowCh <- *newRowResultWithError(err)
			return false
		}

		rowData := make([]interface{}, len(columnTypes))
		for i, ptr := range columnPointers {
			val := reflect.ValueOf(ptr).Elem()
			rowData[i] = val.Interface()
		}
		rowCh <- *newRowResult(rowData)
	}

	if err := queryRows.Err(); err != nil {
		rowCh <- *newRowResultWithError(err)
		return false
	}

	return true
}

func (db *Db) CancelQuery() {
	if db.cancelRunningQuery != nil {
		db.cancelRunningQuery()
	}
}

func treatDbError(originalErr error) error {
	err := originalErr

	if strings.Contains(err.Error(), "interactive transaction not allowed in HTTP queries") {
		err = &shellerrors.TransactionNotSupportedError{}
	}
	if strings.Contains(err.Error(), "context canceled") {
		err = &shellerrors.CancelQueryContextError{}
	}

	return err
}
