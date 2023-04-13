package commands

import (
	"fmt"
	"strings"

	"github.com/libsql/libsql-shell-go/internal/db"
	"github.com/spf13/cobra"
)

var dumpCmd = &cobra.Command{
	Use:   ".dump",
	Short: "Render database content as SQL",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, ok := cmd.Context().Value(dbCtx{}).(*DbCmdConfig)
		if !ok {
			return fmt.Errorf("missing db connection")
		}

		fmt.Fprintln(config.OutF, "PRAGMA foreign_keys=OFF;")

		getTablesStatementResult, err := getDbTables(config)
		if err != nil {
			return err
		}

		err = dumpTables(getTablesStatementResult, config)
		if err != nil {
			return err
		}

		return nil
	},
}

func dumpTables(getTableStatementResult db.StatementResult, config *DbCmdConfig) error {
	for tableNameRowResult := range getTableStatementResult.RowCh {
		if tableNameRowResult.Err != nil {
			return tableNameRowResult.Err
		}
		formattedRow, err := db.FormatData(tableNameRowResult.Row, db.TABLE)
		if err != nil {
			return err
		}

		formattedTableName := formattedRow[0]

		tableInfoStatementResult, err := getTableInfo(config, formattedTableName)
		if err != nil {
			return err
		}

		err = dumpTableInfo(tableInfoStatementResult, config, formattedTableName)
		if err != nil {
			return err
		}

		tableRecordsStatementResult, err := getTableRecords(config, formattedTableName)
		if err != nil {
			return err
		}

		err = dumpTableRecords(tableRecordsStatementResult, config, formattedTableName)
		if err != nil {
			return err
		}

		indexesStatementResult, err := getTableIndexes(config, formattedTableName)
		if err != nil {
			return err
		}

		err = dumpTableIndexes(indexesStatementResult, config, formattedTableName)
		if err != nil {
			return err
		}
	}

	return nil
}

func dumpTableInfo(tableInfoStatementResult db.StatementResult, config *DbCmdConfig, tableName string) error {
	createTableStatement := "CREATE TABLE " + tableName + "("
	var isFirstColumn bool = true

	for tableInfoRowResult := range tableInfoStatementResult.RowCh {
		if tableInfoRowResult.Err != nil {
			return tableInfoRowResult.Err
		}
		if !isFirstColumn {
			createTableStatement += ", "
		}

		fieldName := tableInfoRowResult.Row[1]
		fieldType := tableInfoRowResult.Row[2]
		isNotNull := tableInfoRowResult.Row[3]
		defaultValue := tableInfoRowResult.Row[4]
		isPrimaryKey := tableInfoRowResult.Row[5]

		createTableStatement += fmt.Sprintf("%s %s", fieldName, fieldType)

		if defaultValue != nil {
			createTableStatement += ` DEFAULT ` + fmt.Sprint(defaultValue)
		}

		if isNotNullInt64, ok := isNotNull.(int64); ok {
			if isNotNullInt64 == 1 {
				createTableStatement += " NOT NULL"
			}
		} else if isNotNullFloat64, ok := isNotNull.(float64); ok {
			if isNotNullFloat64 == 1 {
				createTableStatement += " NOT NULL"
			}
		}

		if isPrimaryKeyInt64, ok := isPrimaryKey.(int64); ok {
			if isPrimaryKeyInt64 == 1 {
				createTableStatement += " PRIMARY KEY"
			}
		} else if isPrimaryKeyFloat64, ok := isPrimaryKey.(float64); ok {
			if isPrimaryKeyFloat64 == 1 {
				createTableStatement += " PRIMARY KEY"
			}
		}
		isFirstColumn = false
	}

	createTableStatement += ");"
	fmt.Fprintln(config.OutF, createTableStatement)

	return nil
}

func dumpTableRecords(tableRecordsStatementResult db.StatementResult, config *DbCmdConfig, tableName string) error {
	for tableRecordsRowResult := range tableRecordsStatementResult.RowCh {
		if tableRecordsRowResult.Err != nil {
			return tableRecordsRowResult.Err
		}
		insertStatement := "INSERT INTO " + tableName + " VALUES ("

		tableRecordsFormattedRow, err := db.FormatData(tableRecordsRowResult.Row, db.SQLITE)
		if err != nil {
			return err
		}

		insertStatement += strings.Join(tableRecordsFormattedRow, ", ")
		insertStatement += ");"
		fmt.Fprintln(config.OutF, insertStatement)
	}

	return nil
}

func dumpTableIndexes(indexesStatementResult db.StatementResult, config *DbCmdConfig, tableName string) error {
	for indexesRowResult := range indexesStatementResult.RowCh {
		if indexesRowResult.Err != nil {
			return indexesRowResult.Err
		}

		indexesFormattedRow, err := db.FormatData(indexesRowResult.Row, db.TABLE)
		if err != nil {
			return err
		}

		indexName := indexesFormattedRow[1]
		err = dumpIndex(indexName, config)
		if err != nil {
			return err
		}
	}
	return nil
}

func dumpIndex(indexName string, config *DbCmdConfig) error {
	indexStatementResult, err := getIndex(config, indexName)
	if err != nil {
		return err
	}

	for indexRowResult := range indexStatementResult.RowCh {
		if indexRowResult.Err != nil {
			return indexRowResult.Err
		}

		indexFormattedRow, err := db.FormatData(indexRowResult.Row, db.TABLE)
		if err != nil {
			return err
		}
		index := indexFormattedRow[0]
		fmt.Fprintln(config.OutF, index+";")
	}
	return nil
}

func getDbTables(config *DbCmdConfig) (db.StatementResult, error) {
	listTablesResult, err := config.Db.ExecuteStatements("SELECT name FROM sqlite_master WHERE type='table' and name not like 'sqlite_%' and name != '_litestream_seq' and name != '_litestream_lock' and name != 'libsql_wasm_func_table'")
	if err != nil {
		return db.StatementResult{}, err
	}

	statementResult := <-listTablesResult.StatementResultCh
	if statementResult.Err != nil {
		return db.StatementResult{}, statementResult.Err
	}

	return statementResult, nil
}

func getTableInfo(config *DbCmdConfig, tableName string) (db.StatementResult, error) {
	tableInfoResult, err := config.Db.ExecuteStatements(
		fmt.Sprintf("PRAGMA table_info(%s)", tableName),
	)
	if err != nil {
		return db.StatementResult{}, err
	}

	statementResult := <-tableInfoResult.StatementResultCh
	if statementResult.Err != nil {
		return db.StatementResult{}, statementResult.Err
	}

	return statementResult, nil
}

func getTableRecords(config *DbCmdConfig, tableName string) (db.StatementResult, error) {
	tableRecordsResult, err := config.Db.ExecuteStatements(
		fmt.Sprintf("SELECT * FROM %s", tableName),
	)
	if err != nil {
		return db.StatementResult{}, err
	}

	statementResult := <-tableRecordsResult.StatementResultCh
	if statementResult.Err != nil {
		return db.StatementResult{}, statementResult.Err
	}

	return statementResult, nil
}

func getTableIndexes(config *DbCmdConfig, tableName string) (db.StatementResult, error) {
	tableIndexesResult, err := config.Db.ExecuteStatements(
		fmt.Sprintf("PRAGMA index_list(%s)", tableName),
	)
	if err != nil {
		return db.StatementResult{}, err
	}

	statementResult := <-tableIndexesResult.StatementResultCh
	if statementResult.Err != nil {
		return db.StatementResult{}, statementResult.Err
	}

	return statementResult, nil
}

func getIndex(config *DbCmdConfig, indexName string) (db.StatementResult, error) {
	partialIndexResult, err := config.Db.ExecuteStatements(
		fmt.Sprintf("SELECT REPLACE(REPLACE(sql, ' where ', ' WHERE '), ' on ', ' ON ') AS sql_uppercase FROM sqlite_master WHERE type='index' AND name='%s';", indexName),
	)
	if err != nil {
		return db.StatementResult{}, err
	}
	statementResult := <-partialIndexResult.StatementResultCh
	if statementResult.Err != nil {
		return db.StatementResult{}, statementResult.Err
	}

	return statementResult, nil
}
