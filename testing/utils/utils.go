package utils

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/chiselstrike/libsql-shell/src/lib"
	"github.com/spf13/cobra"
)

func Execute(t *testing.T, c *cobra.Command, args ...string) (string, error) {
	return ExecuteWithInitialInput(t, c, "", args...)
}

func ExecuteWithInitialInput(t *testing.T, c *cobra.Command, initialInput string, args ...string) (string, error) {
	t.Helper()

	buf := new(bytes.Buffer)
	bufIn := new(bytes.Buffer)

	_, err := bufIn.Write([]byte(initialInput))
	if err != nil {
		return "", err
	}

	c.SetOut(buf)
	c.SetErr(buf)
	c.SetIn(bufIn)
	c.SetArgs(args)

	err = c.Execute()
	return strings.TrimSpace(buf.String()), err
}

func GetPrintTableOutput(header []string, data [][]interface{}) string {
	buf := new(bytes.Buffer)

	err := lib.PrintTable(buf, header, data)
	if err != nil {
		return strings.TrimSpace(fmt.Sprintf("%v", err))
	}

	return strings.TrimSpace(buf.String())
}
