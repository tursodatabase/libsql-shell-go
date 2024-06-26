package shell_test

import (
	"fmt"
	"github.com/kirsle/configdir"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/libsql/libsql-shell-go/internal/shell"
	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
)

const historyName = "libsql"

var sharedHistoryFileName = fmt.Sprintf(".%s_shell_history", historyName)

func getExpectedHistoryFullPath(name string) string {
	return fmt.Sprintf("%s/.%s/.%s_shell_history", configdir.LocalConfig("turso"), historyName, name)
}

func TestGetHistoryFileBasedOnMode_GivenLocalHistory_WhenPathIsEmpty_ExpectSharedLocalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := ""
	expectedPath := sharedHistoryFileName
	result := shell.GetHistoryFileBasedOnMode(dbPath, enums.LocalHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenLocalHistory_WhenPathIsValid_ExpectSharedLocalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := "/path/to/my/db.sqlite"
	expectedPath := sharedHistoryFileName
	result := shell.GetHistoryFileBasedOnMode(dbPath, enums.LocalHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenSingleHistory_WhenPathIsValid_ExpectSharedGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := "/path/to/my/db.sqlite"
	expectedPath := getExpectedHistoryFullPath(historyName)
	result := shell.GetHistoryFileBasedOnMode(dbPath, enums.SingleHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenSingleHistory_WhenPathIsEmpty_ExpectSharedGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := ""
	expectedPath := getExpectedHistoryFullPath(historyName)
	result := shell.GetHistoryFileBasedOnMode(dbPath, enums.SingleHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenPerDatabaseHistory_WhenPathIsValid_ExpectSpecificGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := "/path/to/my/db.sqlite"
	expectedPath := getExpectedHistoryFullPath("db")
	result := shell.GetHistoryFileBasedOnMode(dbPath, enums.PerDatabaseHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenPerDatabaseHistory_WhenPathIsEmpty_ExpectSharedGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := ""
	expectedPath := getExpectedHistoryFullPath(historyName)
	result := shell.GetHistoryFileBasedOnMode(dbPath, enums.PerDatabaseHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenPerDatabaseHistory_WhenPathIsHttpUrlWithUser_ExpectSpecificGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := "https://username:password@database-username.domain-name"
	expectedPath := getExpectedHistoryFullPath("database-username.domain-name")
	result := shell.GetHistoryFileBasedOnMode(dbPath, enums.PerDatabaseHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenPerDatabaseHistory_WhenPathIsHttpUrlWithoutUser_ExpectSpecificGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := "https://database-username.domain-name"
	expectedPath := getExpectedHistoryFullPath("database-username.domain-name")
	result := shell.GetHistoryFileBasedOnMode(dbPath, enums.PerDatabaseHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenPerDatabaseHistory_WhenPathIsLibsqlUrl_ExpectSpecificGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := "libsql://database-username.domain-name/?jwt=some_token"
	expectedPath := getExpectedHistoryFullPath("database-username.domain-name")
	result := shell.GetHistoryFileBasedOnMode(dbPath, enums.PerDatabaseHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenPerDatabaseHistory_WhenPathIsWssUrl_ExpectSpecificGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := "wss://database-username.domain-name/?jwt=some_token"
	expectedPath := getExpectedHistoryFullPath("database-username.domain-name")
	result := shell.GetHistoryFileBasedOnMode(dbPath, enums.PerDatabaseHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}

func TestGetHistoryFileBasedOnMode_GivenPerDatabaseHistory_WhenPathIsWsUrl_ExpectSpecificGlobalHistory(t *testing.T) {
	c := qt.New(t)

	dbPath := "ws://database-username.domain-name/?jwt=some_token"
	expectedPath := getExpectedHistoryFullPath("database-username.domain-name")
	result := shell.GetHistoryFileBasedOnMode(dbPath, enums.PerDatabaseHistory, historyName)

	c.Assert(result, qt.Equals, expectedPath)
}
