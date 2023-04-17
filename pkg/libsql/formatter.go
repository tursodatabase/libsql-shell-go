package libsql

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type Formatter interface {
	formatBytes(value []byte) string
	formatString(value string) string
	formatDateTime(value time.Time) string
	formatNull() string
	formatBool(value bool) string
	formatInt(value int64) string
	formatUint(value uint64) string
	formatFloat(value float64) string
}

type FormatType int64

const (
	TABLE FormatType = iota
	SQLITE
	CSV
)

type CommonFormatter struct{}

func (c CommonFormatter) formatNull() string {
	return "NULL"
}

func (c CommonFormatter) formatBool(value bool) string {
	return fmt.Sprintf("%v", value)
}

func (c CommonFormatter) formatInt(value int64) string {
	return fmt.Sprintf("%v", value)
}

func (c CommonFormatter) formatUint(value uint64) string {
	return fmt.Sprintf("%v", value)
}

func (c CommonFormatter) formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

type TableFormatter struct {
	*CommonFormatter
}

func (t TableFormatter) formatBytes(value []byte) string {
	return fmt.Sprintf("0x%X", value)
}

func (t TableFormatter) formatDateTime(value time.Time) string {
	return fmt.Sprint(value.Format("2006-01-02 15:04:05"))
}

func (t TableFormatter) formatString(value string) string {
	return value
}

type SQLiteFormatter struct {
	*CommonFormatter
}

func (s SQLiteFormatter) formatBytes(value []byte) string {
	return fmt.Sprintf("X'%X'", value)
}

func (s SQLiteFormatter) formatDateTime(value time.Time) string {
	return fmt.Sprintf("'%s'", value.Format("2006-01-02 15:04:05"))
}

func (s SQLiteFormatter) formatString(value string) string {
	return fmt.Sprintf("'%v'", value)
}

type CSVFormatter struct {
	*TableFormatter
}

func (c CSVFormatter) formatDateTime(value time.Time) string {
	return fmt.Sprintf("'%s'", value.Format("2006-01-02 15:04:05"))
}

func GetFormatter(format FormatType) Formatter {
	common := &CommonFormatter{}
	switch format {
	case TABLE:
		return &TableFormatter{common}
	case SQLITE:
		return &SQLiteFormatter{common}
	case CSV:
		return CSVFormatter{
			&TableFormatter{common},
		}
	default:
		return nil
	}
}

func FormatData(row []interface{}, format FormatType) ([]string, error) {
	formattedRow := make([]string, len(row))
	formatter := GetFormatter(format)
	for j, val := range row {
		result, err := formatValue(val, formatter)
		if err != nil {
			return nil, err
		}
		formattedRow[j] = result
	}
	return formattedRow, nil
}

func formatValue(val interface{}, formatter Formatter) (string, error) {
	if val == nil {
		return formatter.formatNull(), nil
	} else {
		rv := reflect.ValueOf(val)
		switch rv.Kind() {
		case reflect.Struct:
			return formatStruct(rv, formatter)
		case reflect.Slice:
			return formatSlice(rv, formatter)
		case reflect.Map:
			return formatMap(rv, formatter)
		default:
			formattedRawType, err := formatRawType(rv, formatter)
			if err != nil {
				return "", fmt.Errorf("unsupported type: %s", rv.Kind())
			}
			return formattedRawType, nil
		}
	}
}

func formatStruct(value reflect.Value, formatter Formatter) (string, error) {
	if !value.FieldByName("Valid").IsValid() {
		return "", fmt.Errorf("unsupported struct type: %s. missing Valid field", value.Type().Name())
	}

	if !value.FieldByName("Valid").Bool() {
		return formatter.formatNull(), nil
	}

	switch value.Type().Name() {
	case "NullBool":
		return formatter.formatBool(value.FieldByName("Bool").Bool()), nil
	case "NullByte":
		return formatter.formatInt(value.FieldByName("Byte").Int()), nil
	case "NullInt16":
		return formatter.formatInt(value.FieldByName("Int16").Int()), nil
	case "NullInt32":
		return formatter.formatInt(value.FieldByName("Int16").Int()), nil
	case "NullInt64":
		return formatter.formatInt(value.FieldByName("Int64").Int()), nil
	case "NullFloat64":
		return formatter.formatFloat(value.FieldByName("Float64").Float()), nil
	case "NullString":
		return formatter.formatString(value.FieldByName("String").String()), nil
	case "NullTime":
		return formatter.formatDateTime(value.FieldByName("Time").Interface().(time.Time)), nil
	default:
		return "", fmt.Errorf("unsupported struct type: %s", value.Type().Name())
	}
}

func formatSlice(value reflect.Value, formatter Formatter) (string, error) {
	if value.Type().Elem().Kind() == reflect.Uint8 {
		return formatter.formatBytes(value.Interface().([]byte)), nil
	}

	return "", fmt.Errorf("unsupported slice: %s", value.Type().Name())
}

func formatMap(value reflect.Value, formatter Formatter) (string, error) {
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

	sliceOfBytesValue, err := decodeBase64(base64ValueString)
	if err != nil {
		return "", err
	}

	return formatter.formatBytes(sliceOfBytesValue), nil
}

func decodeBase64(base64String string) ([]byte, error) {
	encodingWithNoPadding := base64.StdEncoding.WithPadding(base64.NoPadding)
	decodedBase64 := make([]byte, encodingWithNoPadding.DecodedLen(len(base64String)))
	_, err := encodingWithNoPadding.Decode(decodedBase64, []byte(base64String))
	if err != nil {
		return []byte{}, errors.Join(errors.New("unable to decode base64 value"), err)
	}
	return decodedBase64, nil
}

func formatRawType(value reflect.Value, formatter Formatter) (string, error) {
	switch value.Kind() {
	case reflect.Bool:
		return formatter.formatBool(value.Bool()), nil
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		return formatter.formatInt(value.Int()), nil
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		return formatter.formatUint(value.Uint()), nil
	case reflect.String:
		return formatter.formatString(value.String()), nil
	case reflect.Float32,
		reflect.Float64:
		return formatter.formatFloat(value.Float()), nil
	default:
		return "", fmt.Errorf("unsupported raw type: %s", value.Kind())
	}
}
