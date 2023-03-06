package lib

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
)

func PrintStatementsResult(result Result, outF io.Writer, withoutHeader bool) error {
	if len(result.ColumnNames) == 0 {
		return nil
	}

	table := createTable(outF)
	if !withoutHeader {
		table.SetHeader(result.ColumnNames)
	}

	for row := range result.RowCh {
		if row.Err != nil {
			return row.Err
		}
		formattedRow, err := formatData(row.Row)

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

func formatData(row []interface{}) ([]string, error) {
	formattedRow := make([]string, len(row))
	for j, val := range row {
		result, err := formatValue(val)
		if err != nil {
			return nil, err
		}
		formattedRow[j] = result
	}
	return formattedRow, nil
}

func formatValue(val interface{}) (string, error) {
	if val == nil {
		return "NULL", nil
	} else {
		rv := reflect.ValueOf(val)

		switch rv.Kind() {
		case reflect.Struct:
			return formatStruct(rv)
		case reflect.Slice:
			return formatSlice(rv)
		default:
			formattedRawType, err := formatRawTypes(rv)
			if err != nil {
				return "", fmt.Errorf("unsupported type: %s", rv.Kind())
			}
			return formattedRawType, nil
		}
	}
}

func formatStruct(value reflect.Value) (string, error) {
	if !value.FieldByName("Valid").IsValid() {
		return "", fmt.Errorf("unsupported struct type: %s. missing Valid field", value.Type().Name())
	}

	if !value.FieldByName("Valid").Bool() {
		return "NULL", nil
	}

	switch value.Type().Name() {
	case "NullBool":
		return formatRawTypes(value.FieldByName("Bool"))
	case "NullFloat64":
		return formatRawTypes(value.FieldByName("Float64"))
	case "NullByte":
		return formatRawTypes(value.FieldByName("Byte"))
	case "NullInt16":
		return formatRawTypes(value.FieldByName("Int16"))
	case "NullInt32":
		return formatRawTypes(value.FieldByName("Int32"))
	case "NullInt64":
		return formatRawTypes(value.FieldByName("Int64"))
	case "NullString":
		return formatRawTypes(value.FieldByName("String"))
	case "NullTime":
		return value.FieldByName("Time").Interface().(time.Time).Format("2006-01-02 15:04:05"), nil
	default:
		return "", fmt.Errorf("unsupported struct type: %s", value.Type().Name())
	}
}

func formatSlice(value reflect.Value) (string, error) {
	if value.Type().Elem().Kind() == reflect.Uint8 {
		return fmt.Sprintf("%x", value.Interface()), nil
	}

	return "", fmt.Errorf("unsupported slice: %s", value.Type().Name())
}

func formatRawTypes(value reflect.Value) (string, error) {
	switch value.Kind() {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.String:
		contentValue := value.Interface()
		return fmt.Sprintf("%v", contentValue), nil
	case reflect.Float32,
		reflect.Float64:
		contentValue := value.Float()
		return strconv.FormatFloat(contentValue, 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("unsupported raw type: %s", value.Kind())
	}
}
