package libsql

import (
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
)

func PrintStatementsResult(statementsResult StatementsResult, outF io.Writer, withoutHeader bool) error {
	if statementsResult.StatementResultCh == nil {
		return &InvalidStatementsResult{}
	}

	for statementResult := range statementsResult.StatementResultCh {
		if statementResult.Err != nil {
			return statementResult.Err
		}

		err := PrintStatementResult(statementResult, outF, withoutHeader)
		if err != nil {
			return err
		}
	}
	return nil
}

func PrintStatementResult(statementResult StatementResult, outF io.Writer, withoutHeader bool) error {
	if statementResult.RowCh == nil {
		return &UnableToPrintStatementResult{}
	}

	if len(statementResult.ColumnNames) == 0 {
		return nil
	}

	table := createTable(outF)
	if !withoutHeader {
		table.SetHeader(statementResult.ColumnNames)
	}

	for row := range statementResult.RowCh {
		if row.Err != nil {
			return row.Err
		}
		formattedRow, err := FormatData(row.Row, TABLE)

		if err != nil {
			return err
		}
		table.Append(formattedRow)
	}

	table.Render()
	return nil
}

func PrintError(err error, errF io.Writer) {
	fmt.Fprintf(errF, "Error: %s\n", err.Error())
}

func createTable(outF io.Writer) *tablewriter.Table {
	table := tablewriter.NewWriter(outF)

	table.SetHeaderLine(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoFormatHeaders(true)

	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColumnSeparator("  ")
	table.SetNoWhiteSpace(true)
	table.SetTablePadding("     ")

	return table
}

func PrintTable(outF io.Writer, header []string, data [][]string) {
	table := createTable(outF)
	table.SetHeader(header)
	table.AppendBulk(data)
	table.Render()
}
