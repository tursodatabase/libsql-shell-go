package lib

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const QUIT_COMMAND = ".quit"
const DEFAULT_WELCOME_MESSAGE = "Welcome to LibSQL shell!\n\nType \".quit\" to exit the shell, \".tables\" to list all tables, and \".schema\" to show table schemas.\n\n"

const promptNewStatement = "→  "
const promptContinueStatement = "... "

type ShellConfig struct {
	InF            io.Reader
	OutF           io.Writer
	ErrF           io.Writer
	HistoryMode    HistoryMode
	HistoryName    string
	QuietMode      bool
	WelcomeMessage *string
}

type shell struct {
	config ShellConfig

	db        *Db
	promptFmt func(p ...interface{}) string

	readline                 *readline.Instance
	statementParts           []string
	insideMultilineStatement bool
}

func (db *Db) RunShell(config ShellConfig) error {
	shellInstance, err := newShell(config, db)
	if err != nil {
		return err
	}
	return shellInstance.run()
}

func newShell(config ShellConfig, db *Db) (*shell, error) {
	promptFmt := color.New(color.FgBlue, color.Bold).SprintFunc()
	return &shell{config: config, db: db, promptFmt: promptFmt}, nil
}

func (sh *shell) run() error {
	var err error
	sh.readline, err = sh.newReadline()
	if err != nil {
		return err
	}
	defer sh.readline.Close()
	sh.readline.CaptureExitSignal()

	if !sh.config.QuietMode {
		fmt.Print(sh.getWelcomeMessage())
	}

	dbCmdConfig := &dbCmdConfig{
		db:   sh.db,
		OutF: sh.config.OutF,
		ErrF: sh.config.ErrF,
	}

	databaseCmd := CreateNewDatabaseRootCmd(dbCmdConfig)

	for {
		line, err := sh.readline.Readline()

		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				return nil
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)

		switch {
		case len(line) == 0:
			continue
		case sh.insideMultilineStatement:
			sh.appendStatementPartAndExecuteIfFinished(line)
		case line == QUIT_COMMAND:
			return nil
		case isCommand(line):
			err = sh.executeCommand(databaseCmd, line)
			if err != nil {
				rx := regexp.MustCompile(`"[^"]*"`)
				command := rx.FindString(fmt.Sprint(err))
				if command == "" {
					PrintError(fmt.Errorf(`unknown command or invalid arguments. Enter ".help" for help`), dbCmdConfig.ErrF)
				}
				errorMsg := fmt.Sprintf(`unknown command or invalid arguments: %s. Enter ".help" for help`, command)
				PrintError(fmt.Errorf(errorMsg), dbCmdConfig.ErrF)
			}
		default:
			sh.appendStatementPartAndExecuteIfFinished(line)
		}

	}
	return nil
}

func (sh *shell) newReadline() (*readline.Instance, error) {
	historyFile := GetHistoryFileBasedOnMode(sh.db.path, sh.config.HistoryMode, sh.config.HistoryName)

	return readline.NewEx(&readline.Config{
		Prompt:          sh.promptFmt(promptNewStatement),
		InterruptPrompt: "^C",
		HistoryFile:     historyFile,
		EOFPrompt:       QUIT_COMMAND,
		Stdin:           io.NopCloser(sh.config.InF),
		Stdout:          sh.config.OutF,
		Stderr:          sh.config.ErrF,
	})
}

func isCommand(line string) bool {
	return line[0] == '.'
}

func (sh *shell) executeCommand(databaseCmd *cobra.Command, command string) error {
	parts := strings.Fields(command)
	databaseCmd.SetArgs(parts)

	err := databaseCmd.Execute()
	if err != nil {
		return err
	}
	return nil
}

func (sh *shell) appendStatementPartAndExecuteIfFinished(statementPart string) {
	sh.statementParts = append(sh.statementParts, statementPart)
	if strings.HasSuffix(statementPart, ";") {
		completeStatement := strings.Join(sh.statementParts, "\n")
		sh.statementParts = make([]string, 0)
		sh.insideMultilineStatement = false
		sh.readline.SetPrompt(sh.promptFmt(promptNewStatement))
		sh.db.ExecuteAndPrintStatements(completeStatement, sh.readline.Stdout(), sh.readline.Stderr(), false)
	} else {
		sh.readline.SetPrompt(sh.promptFmt(promptContinueStatement))
		sh.insideMultilineStatement = false
	}
}

func (sh *shell) getWelcomeMessage() string {
	if sh.config.WelcomeMessage == nil {
		return DEFAULT_WELCOME_MESSAGE
	}
	return *sh.config.WelcomeMessage
}
