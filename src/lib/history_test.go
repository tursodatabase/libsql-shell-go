package lib_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/chiselstrike/libsql-shell/src/lib"
	qt "github.com/frankban/quicktest"
)

const historyName = "libsql"

var sharedHistoryFileName = fmt.Sprintf(".%s_shell_history", historyName)

func getExpectedHistoryFullPath(name string) string {
	return fmt.Sprintf("%s/.%s/.%s_shell_history", os.Getenv("HOME"), historyName, name)
}

func TestGetHistoryFileBasedOnMode_GivenLocalHistory_WhenPathIsEmpty_ExpectSharedLocalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := ""
	expectedPath := sharedHistoryFileName
	result := lib.GetHistoryFileBasedOnMode(dbPath, lib.LocalHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenLocalHistory_WhenPathIsValid_ExpectSharedLocalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := "/path/to/my/db.sqlite"
	expectedPath := sharedHistoryFileName
	result := lib.GetHistoryFileBasedOnMode(dbPath, lib.LocalHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenSingleHistory_WhenPathIsValid_ExpectSharedGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := "/path/to/my/db.sqlite"
	expectedPath := getExpectedHistoryFullPath(historyName)
	result := lib.GetHistoryFileBasedOnMode(dbPath, lib.SingleHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenSingleHistory_WhenPathIsEmpty_ExpectSharedGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := ""
	expectedPath := getExpectedHistoryFullPath(historyName)
	result := lib.GetHistoryFileBasedOnMode(dbPath, lib.SingleHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenPerDatabaseHistory_WhenPathIsValid_ExpectSpecificGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := "/path/to/my/db.sqlite"
	expectedPath := getExpectedHistoryFullPath("db")
	result := lib.GetHistoryFileBasedOnMode(dbPath, lib.PerDatabaseHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenPerDatabaseHistory_WhenPathIsEmpty_ExpectSharedGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := ""
	expectedPath := getExpectedHistoryFullPath(historyName)
	result := lib.GetHistoryFileBasedOnMode(dbPath, lib.PerDatabaseHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenPerDatabaseHistory_WhenPathIsHttpUrl_ExpectSpecificGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := "https://username:password@company.turso.io"
	expectedPath := getExpectedHistoryFullPath("username:password")
	result := lib.GetHistoryFileBasedOnMode(dbPath, lib.PerDatabaseHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenPerDatabaseHistory_WhenPathIsHttpUrlWithoutUser_ExpectSharedGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := "https://company.turso.io"
	expectedPath := getExpectedHistoryFullPath(historyName)
	result := lib.GetHistoryFileBasedOnMode(dbPath, lib.PerDatabaseHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}
