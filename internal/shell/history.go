package shell

import (
	"fmt"
	"github.com/kirsle/configdir"
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
		if host, err := getHostFromDbUri(dbPath); err == nil && host != "" {
			return getHistoryFileFullPath(historyName, getHistoryFileName(host))
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
	path := filepath.Join(configdir.LocalConfig("turso"), fmt.Sprintf(".%s", historyName))
	_ = os.MkdirAll(path, os.ModePerm)
	return path
}

func getHostFromDbUri(dbPath string) (string, error) {
	if db.IsUrl(dbPath) {
		url, err := url.Parse(dbPath)
		if err != nil {
			return "", err
		}
		if url.Host == "" {
			return "", &shellerrors.UrlDoesNotContainHostError{}
		}
		return url.Host, nil
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
