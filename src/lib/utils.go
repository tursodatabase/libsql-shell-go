package lib

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

func PrintTable(outF io.Writer, header []string, data [][]interface{}) error {
	table := tablewriter.NewWriter(outF)

	formattedData, err := formatData(data)
	if err != nil {
		return err
	}
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

	table.AppendBulk(formattedData)

	table.Render()

	return nil
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
		hexString := hex.EncodeToString(value.Bytes())
		base64Str := strings.TrimRight(base64.StdEncoding.EncodeToString(value.Bytes()), "=")
		sliceOfBytes := make([]byte, base64.StdEncoding.DecodedLen(len(hexString)))

		_, err := base64.StdEncoding.Decode(sliceOfBytes, []byte(base64Str))
		if err != nil {
			return base64Str, nil
		}
		return hexString, nil
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
