package lib

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

func PrintTable(outF io.Writer, header []string, data [][]string) {
	table := tablewriter.NewWriter(outF)

	table.SetHeader(header)
	table.SetHeaderLine(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoFormatHeaders(true)

	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColumnSeparator("  ")
	table.SetNoWhiteSpace(true)
	table.SetTablePadding("     ")

	table.AppendBulk(data)

	table.Render()
}
