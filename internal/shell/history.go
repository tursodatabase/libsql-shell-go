package shell

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/libsql/libsql-shell-go/internal/db"
	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
	"github.com/libsql/libsql-shell-go/pkg/shell/shellerrors"
)

func GetHistoryFileBasedOnMode(dbPath string, mode enums.HistoryMode, historyName string) string {
	sharedHistoryFileName := getHistoryFileName(historyName)

	switch mode {
	case enums.LocalHistory:
		return sharedHistoryFileName
	case enums.PerDatabaseHistory:
		if parsedName, err := parseNameFromDbPath(dbPath); err == nil && parsedName != "" {
			return getHistoryFileFullPath(historyName, getHistoryFileName(parsedName))
		}
	}

	return getHistoryFileFullPath(historyName, sharedHistoryFileName)
}

func getHistoryFileFullPath(historyName string, fileName string) string {
	return filepath.Join(getHistoryFolderPath(historyName), fileName)
}

func getHistoryFileName(name string) string {
	return fmt.Sprintf(".%s_shell_history", name)
}

func getHistoryFolderPath(historyName string) string {
	path := filepath.Join(os.Getenv("HOME"), fmt.Sprintf(".%s", historyName))
	_ = os.MkdirAll(path, os.ModePerm)
	return path
}

func parseNameFromDbPath(dbPath string) (string, error) {
	if db.IsUrl(dbPath) {
		url, err := url.Parse(dbPath)
		if err != nil {
			return "", err
		}
		if url.User == nil {
			return "", &shellerrors.UrlDoesNotContainUserError{}
		}
		return url.User.String(), nil
	}

	return getFileNameWithoutExtension(dbPath), nil
}

func getFileNameWithoutExtension(path string) string {
	filename := filepath.Base(path)
	extension := filepath.Ext(path)
	filenameWithoutExtension := filename[0 : len(filename)-len(extension)]

	if filenameWithoutExtension == "." {
		return ""
	}
	return filenameWithoutExtension
}
