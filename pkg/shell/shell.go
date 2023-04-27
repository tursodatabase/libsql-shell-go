package shell

import (
	"io"

	"github.com/libsql/libsql-shell-go/internal/db"
	"github.com/libsql/libsql-shell-go/internal/shell"
	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
)

type ShellConfig struct {
	DbPath                    string
	InF                       io.Reader
	OutF                      io.Writer
	ErrF                      io.Writer
	HistoryMode               enums.HistoryMode
	HistoryName               string
	QuietMode                 bool
	WelcomeMessage            *string
	AfterDbConnectionCallback func()
}

func RunShell(config ShellConfig) error {
	db, err := db.NewDb(config.DbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if config.AfterDbConnectionCallback != nil {
		config.AfterDbConnectionCallback()
	}

	internalConfig := publicToInternalConfig(config)
	shellInstance, err := shell.NewShell(internalConfig, db)
	if err != nil {
		return err
	}
	return shellInstance.Run()
}

func RunShellLine(config ShellConfig, line string) error {
	db, err := db.NewDb(config.DbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if config.AfterDbConnectionCallback != nil {
		config.AfterDbConnectionCallback()
	}

	internalConfig := publicToInternalConfig(config)
	shellInstance, err := shell.NewShell(internalConfig, db)
	if err != nil {
		return err
	}

	return shellInstance.ExecuteCommandOrStatements(line)
}

func publicToInternalConfig(publicConfig ShellConfig) shell.ShellConfig {
	return shell.ShellConfig{
		InF:            publicConfig.InF,
		OutF:           publicConfig.OutF,
		ErrF:           publicConfig.ErrF,
		HistoryMode:    publicConfig.HistoryMode,
		HistoryName:    publicConfig.HistoryName,
		QuietMode:      publicConfig.QuietMode,
		WelcomeMessage: publicConfig.WelcomeMessage,
	}
}
