package shell

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/libsql/libsql-shell-go/pkg/libsql"
	"github.com/libsql/libsql-shell-go/shell/commands"
	"github.com/spf13/cobra"
)

const QUIT_COMMAND = ".quit"
const DEFAULT_WELCOME_MESSAGE = "Welcome to LibSQL shell!\n\nType \".quit\" to exit the shell, and \".help\" to show all commands\n\n"

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

	db        *libsql.Db
	promptFmt func(p ...interface{}) string

	readline                 *readline.Instance
	statementParts           []string
	insideMultilineStatement bool

	databaseCmd *cobra.Command
}

func RunShell(db *libsql.Db, config ShellConfig) error {
	shellInstance, err := newShell(config, db)
	if err != nil {
		return err
	}
	return shellInstance.run()
}

func RunShellCommandOrStatements(db *libsql.Db, config ShellConfig, commandOrStatements string) error {
	shellInstance, err := newShell(config, db)
	if err != nil {
		return err
	}
	return shellInstance.executeCommandOrStatements(commandOrStatements)
}

func newShell(config ShellConfig, db *libsql.Db) (*shell, error) {
	promptFmt := color.New(color.FgBlue, color.Bold).SprintFunc()

	dbCmdConfig := &commands.DbCmdConfig{
		Db:   db,
		OutF: config.OutF,
		ErrF: config.ErrF,
	}
	databaseCmd := commands.CreateNewDatabaseRootCmd(dbCmdConfig)

	return &shell{config: config, db: db, promptFmt: promptFmt, databaseCmd: databaseCmd}, nil
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
			err = sh.executeCommand(line)
			if err != nil {
				libsql.PrintError(err, sh.config.ErrF)
			}
		default:
			sh.appendStatementPartAndExecuteIfFinished(line)
		}

	}
	return nil
}

func (sh *shell) newReadline() (*readline.Instance, error) {
	historyFile := GetHistoryFileBasedOnMode(sh.db.Path, sh.config.HistoryMode, sh.config.HistoryName)

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

func (sh *shell) executeCommand(command string) error {
	parts := strings.Fields(command)
	sh.databaseCmd.SetArgs(parts)

	err := sh.databaseCmd.Execute()

	if err != nil && strings.HasPrefix(err.Error(), "unknown command") {
		rx := regexp.MustCompile(`"[^"]*"`)
		command := rx.FindString(fmt.Sprint(err))
		return fmt.Errorf(`unknown command or invalid arguments: %s. Enter ".help" for help`, command)
	}
	return err
}

func (sh *shell) appendStatementPartAndExecuteIfFinished(statementPart string) {
	sh.statementParts = append(sh.statementParts, statementPart)
	if strings.HasSuffix(statementPart, ";") {
		completeStatement := strings.Join(sh.statementParts, "\n")
		sh.statementParts = make([]string, 0)
		sh.insideMultilineStatement = false
		sh.readline.SetPrompt(sh.promptFmt(promptNewStatement))
		err := sh.db.ExecuteAndPrintStatements(completeStatement, sh.readline.Stdout(), false)
		if err != nil {
			libsql.PrintError(err, sh.readline.Stderr())
		}
	} else {
		sh.readline.SetPrompt(sh.promptFmt(promptContinueStatement))
		sh.insideMultilineStatement = false
	}
}

func (sh *shell) executeCommandOrStatements(commandOrStatements string) error {
	if isCommand(commandOrStatements) {
		return sh.executeCommand(commandOrStatements)
	}

	return sh.db.ExecuteAndPrintStatements(commandOrStatements, sh.config.OutF, false)
}

func (sh *shell) getWelcomeMessage() string {
	if sh.config.WelcomeMessage == nil {
		return DEFAULT_WELCOME_MESSAGE
	}
	return *sh.config.WelcomeMessage
}
