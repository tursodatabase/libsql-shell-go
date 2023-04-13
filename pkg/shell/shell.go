package shell

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/libsql/libsql-shell-go/internal/db"
	"github.com/libsql/libsql-shell-go/pkg/shell/commands"
	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
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

	db        *db.Db
	promptFmt func(p ...interface{}) string

	state shellState

	databaseCmd *cobra.Command
}

type shellState struct {
	readline                   *readline.Instance
	statementParts             []string
	insideMultilineStatement   bool
	interruptReadEvalPrintLoop bool
	printMode                  enums.PrintMode
}

func RunShell(db *db.Db, config ShellConfig) error {
	shellInstance, err := newShell(config, db)
	if err != nil {
		return err
	}
	return shellInstance.run()
}

func RunShellCommandOrStatements(db *db.Db, config ShellConfig, commandOrStatements string) error {
	shellInstance, err := newShell(config, db)
	if err != nil {
		return err
	}

	err = shellInstance.resetState()
	if err != nil {
		return err
	}
	return shellInstance.executeCommandOrStatements(commandOrStatements)
}

func newShell(config ShellConfig, db *db.Db) (*shell, error) {
	promptFmt := color.New(color.FgBlue, color.Bold).SprintFunc()

	newShell := shell{config: config, db: db, promptFmt: promptFmt}

	dbCmdConfig := &commands.DbCmdConfig{
		Db:                db,
		OutF:              config.OutF,
		ErrF:              config.ErrF,
		SetInterruptShell: func() { newShell.state.interruptReadEvalPrintLoop = true },
		SetMode:           func(mode enums.PrintMode) { newShell.state.printMode = mode },
		GetMode: func() enums.PrintMode {
			return newShell.state.printMode
		},
	}
	newShell.databaseCmd = commands.CreateNewDatabaseRootCmd(dbCmdConfig)

	return &newShell, nil
}

func (sh *shell) run() error {
	err := sh.resetState()
	if err != nil {
		return err
	}
	defer sh.state.readline.Close()

	if !sh.config.QuietMode {
		fmt.Print(sh.getWelcomeMessage())
	}

	for !sh.state.interruptReadEvalPrintLoop {
		line, err := sh.state.readline.Readline()

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
		case sh.state.insideMultilineStatement:
			sh.appendStatementPartAndExecuteIfFinished(line)
		case isCommand(line):
			err = sh.executeCommand(line)
			if err != nil {
				db.PrintError(err, sh.config.ErrF)
			}
		default:
			sh.appendStatementPartAndExecuteIfFinished(line)
		}

	}
	return nil
}

func (sh *shell) resetState() error {
	var err error
	sh.state.readline, err = sh.newReadline()
	if err != nil {
		return err
	}
	sh.state.readline.CaptureExitSignal()

	sh.state.insideMultilineStatement = false
	sh.state.statementParts = make([]string, 0)

	sh.state.interruptReadEvalPrintLoop = false

	sh.state.printMode = enums.TABLE_MODE

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
	sh.state.statementParts = append(sh.state.statementParts, statementPart)
	if strings.HasSuffix(statementPart, ";") {
		completeStatement := strings.Join(sh.state.statementParts, "\n")
		sh.state.statementParts = make([]string, 0)
		sh.state.insideMultilineStatement = false
		sh.state.readline.SetPrompt(sh.promptFmt(promptNewStatement))
		err := sh.db.ExecuteAndPrintStatements(completeStatement, sh.config.OutF, false, sh.state.printMode)
		if err != nil {
			db.PrintError(err, sh.state.readline.Stderr())
		}
	} else {
		sh.state.readline.SetPrompt(sh.promptFmt(promptContinueStatement))
		sh.state.insideMultilineStatement = true
	}
}

func (sh *shell) executeCommandOrStatements(commandOrStatements string) error {
	if isCommand(commandOrStatements) {
		return sh.executeCommand(commandOrStatements)
	}

	return sh.db.ExecuteAndPrintStatements(commandOrStatements, sh.config.OutF, false, sh.state.printMode)
}

func (sh *shell) getWelcomeMessage() string {
	if sh.config.WelcomeMessage == nil {
		return DEFAULT_WELCOME_MESSAGE
	}
	return *sh.config.WelcomeMessage
}
