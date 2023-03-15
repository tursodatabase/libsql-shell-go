package utils

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chiselstrike/libsql-shell/pkg/libsql"
	"github.com/spf13/cobra"
)

func Execute(t *testing.T, c *cobra.Command, args ...string) (string, string, error) {
	return ExecuteWithInitialInput(t, c, "", args...)
}

func ExecuteWithInitialInput(t *testing.T, c *cobra.Command, initialInput string, args ...string) (string, string, error) {
	t.Helper()

	bufOut := new(bytes.Buffer)
	bufErr := new(bytes.Buffer)
	bufIn := new(bytes.Buffer)

	_, err := bufIn.Write([]byte(initialInput))
	if err != nil {
		return "", "", err
	}

	c.SetOut(bufOut)
	c.SetErr(bufErr)
	c.SetIn(bufIn)
	c.SetArgs(args)

	err = c.Execute()
	return strings.TrimSpace(bufOut.String()), strings.TrimSpace(bufErr.String()), err
}

func GetPrintTableOutput(header []string, data [][]string) string {
	buf := new(bytes.Buffer)

	libsql.PrintTable(buf, header, data)

	return strings.TrimSpace(buf.String())
}
