package shell

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/libsql/libsql-shell-go/internal/db"
)

type HistoryMode int64

const (
	SingleHistory      HistoryMode = 0
	PerDatabaseHistory HistoryMode = 1
	LocalHistory       HistoryMode = 2
)

func GetHistoryFileBasedOnMode(dbPath string, mode HistoryMode, historyName string) string {
	sharedHistoryFileName := getHistoryFileName(historyName)

	switch mode {
	case LocalHistory:
		return sharedHistoryFileName
	case PerDatabaseHistory:
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
	if db.IsHttpUrl(dbPath) {
		url, err := url.Parse(dbPath)
		if err != nil {
			return "", err
		}
		if url.User == nil {
			return "", &db.UrlDoesNotContainUserError{}
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
