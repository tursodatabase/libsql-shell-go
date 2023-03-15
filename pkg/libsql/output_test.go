package libsql_test

import (
	"testing"

	"github.com/chiselstrike/libsql-shell/testing/utils"
	qt "github.com/frankban/quicktest"
)

func TestGetTableOutput_GivenHeaderWithoutData_ExpectTableHasJustHeader(t *testing.T) {
	c := qt.New(t)

	header := []string{"id", "value"}
	data := [][]string{}
	result := utils.GetPrintTableOutput(header, data)

	c.Assert(result, qt.Equals, "ID     VALUE")
}

func TestGetTableOutput_GivenHeaderWithData_ExpectTableHasHeaderAndData(t *testing.T) {
	c := qt.New(t)

	header := []string{"id", "value"}
	data := [][]string{{"1", "test"}}
	result := utils.GetPrintTableOutput(header, data)

	c.Assert(result, qt.Equals, "ID     VALUE \n1      test")
}

func TestGetTableOutput_GivenDataWithoutHeader_ExpectTableHasJustData(t *testing.T) {
	c := qt.New(t)

	header := []string{}
	data := [][]string{{"1", "test"}}
	result := utils.GetPrintTableOutput(header, data)

	c.Assert(result, qt.Equals, "1     test")
}

func TestGetTableOutput_GivenHeaderWithMultipleRows_ExpectTableHasHeaderAndData(t *testing.T) {
	c := qt.New(t)

	header := []string{"id", "value"}
	data := [][]string{{"1", "test"}, {"2", "test2"}}
	result := utils.GetPrintTableOutput(header, data)

	c.Assert(result, qt.Equals, "ID     VALUE \n1      test      \n2      test2")
}

func TestGetTableOutput_GivenHeaderWithMultipleRowsAndDifferentLength_ExpectTableHasHeaderAndData(t *testing.T) {
	c := qt.New(t)

	header := []string{"id", "value"}
	data := [][]string{{"1", "test"}, {"2", "test2", "test3"}}
	result := utils.GetPrintTableOutput(header, data)

	c.Assert(result, qt.Equals, "ID     VALUE           \n1      test      \n2      test2     test3")
}

func TestGetTableOutput_GivenHeaderNilAndData_ExpectTableHasJustData(t *testing.T) {
	c := qt.New(t)

	header := []string(nil)
	data := [][]string{{"1", "test"}}
	result := utils.GetPrintTableOutput(header, data)

	c.Assert(result, qt.Equals, "1     test")
}

func TestGetTableOutput_GivenHeaderAndDataNil_ExpectTableHasJustHeader(t *testing.T) {
	c := qt.New(t)

	header := []string{"id", "value"}
	data := [][]string(nil)
	result := utils.GetPrintTableOutput(header, data)

	c.Assert(result, qt.Equals, "ID     VALUE")
}
