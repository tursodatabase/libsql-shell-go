package lib

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
)

func PrintStatementsResults(results []Result, outF io.Writer, withoutHeader bool) error {
	for _, result := range results {
		if len(result.ColumnNames) != 0 {
			formattedData, err := formatData(result.Data)
			if err != nil {
				return err
			}

			if withoutHeader {
				PrintTable(outF, nil, formattedData)
			} else {
				PrintTable(outF, result.ColumnNames, formattedData)
			}
		}
	}
	return nil
}

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

func formatData(data [][]interface{}) ([][]string, error) {
	formattedData := make([][]string, len(data))
	for i, row := range data {
		formattedRow := make([]string, len(row))
		for j, val := range row {
			result, err := formatValue(val)
			if err != nil {
				return nil, err
			}
			formattedRow[j] = result
		}
		formattedData[i] = formattedRow
	}
	return formattedData, nil
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
		case reflect.Map:
			return formatMap(rv)
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
		return formatBytes(value.Interface().([]byte)), nil
	}

	return "", fmt.Errorf("unsupported slice: %s", value.Type().Name())
}

func formatMap(value reflect.Value) (string, error) {
	base64Value := value.MapIndex(reflect.ValueOf("base64"))
	if base64Value.IsZero() {
		return "", fmt.Errorf("unsupported map: no \"base64\" field")
	}

	var base64ValueString string
	switch {
	case base64Value.Kind() == reflect.Interface && base64Value.Elem().Kind() == reflect.String:
		base64ValueString = base64Value.Elem().String()
	case base64Value.Kind() == reflect.String:
		base64ValueString = base64Value.String()
	default:
		return "", fmt.Errorf("unsupported map. unsupported \"base64\" field kind")
	}

	return decodeBase64ToHex(base64ValueString)
}

func decodeBase64ToHex(base64String string) (string, error) {
	encodingWithNoPadding := base64.StdEncoding.WithPadding(base64.NoPadding)

	decodedBase64 := make([]byte, encodingWithNoPadding.DecodedLen(len(base64String)))
	_, err := encodingWithNoPadding.Decode(decodedBase64, []byte(base64String))
	if err != nil {
		return "", errors.Join(errors.New("unable to decode base64 value"), err)
	}

	return formatBytes(decodedBase64), nil
}

func formatBytes(bytes []byte) string {
	return fmt.Sprintf("0x%X", bytes)
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
