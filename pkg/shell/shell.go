package shell

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/libsql/libsql-shell-go/internal/db"
	"github.com/libsql/libsql-shell-go/internal/shell"
	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
)

type ShellConfig struct {
	DbUri                     string
	AuthToken                 string
	InF                       io.Reader
	OutF                      io.Writer
	ErrF                      io.Writer
	HistoryMode               enums.HistoryMode
	HistoryName               string
	QuietMode                 bool
	WelcomeMessage            *string
	AfterDbConnectionCallback func()
	DisableAutoCompletion     bool
}

func RunShell(config ShellConfig) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	db, err := db.NewDb(config.DbUri, config.AuthToken)
	if err != nil {
		return err
	}
	if err := db.TestConnection(); err != nil {
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

	go func() {
		for range signals {
			shellInstance.CancelQuery()
		}
	}()
	return shellInstance.Run()
}

func RunShellLine(config ShellConfig, line string) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	db, err := db.NewDb(config.DbUri, config.AuthToken)
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

	go func() {
		<-signals
		shellInstance.CancelQuery()
	}()

	return shellInstance.ExecuteCommandOrStatements(line)
}

func publicToInternalConfig(publicConfig ShellConfig) shell.ShellConfig {
	return shell.ShellConfig{
		InF:                   publicConfig.InF,
		OutF:                  publicConfig.OutF,
		ErrF:                  publicConfig.ErrF,
		HistoryMode:           publicConfig.HistoryMode,
		HistoryName:           publicConfig.HistoryName,
		QuietMode:             publicConfig.QuietMode,
		WelcomeMessage:        publicConfig.WelcomeMessage,
		DisableAutoCompletion: publicConfig.DisableAutoCompletion,
	}
}
