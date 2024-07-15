package shell

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/libsql/libsql-shell-go/internal/db"
	"github.com/libsql/libsql-shell-go/internal/shellcmd"
	"github.com/libsql/libsql-shell-go/internal/suggester"
	"github.com/libsql/libsql-shell-go/pkg/shell/enums"
	"github.com/spf13/cobra"
	"github.com/tursodatabase/libsql-client-go/sqliteparser"
	"github.com/tursodatabase/libsql-client-go/sqliteparserutils"
)

const QUIT_COMMAND = ".quit"
const DEFAULT_WELCOME_MESSAGE = "Welcome to LibSQL shell!\n\nType \".quit\" to exit the shell, and \".help\" to show all commands\n\n"

const promptNewStatement = "→  "
const promptContinueStatement = "... "

type ShellConfig struct {
	InF                   io.Reader
	OutF                  io.Writer
	ErrF                  io.Writer
	HistoryMode           enums.HistoryMode
	HistoryName           string
	QuietMode             bool
	WelcomeMessage        *string
	DisableAutoCompletion bool
}

type Shell struct {
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

func NewShell(config ShellConfig, db *db.Db) (*Shell, error) {
	promptFmt := color.New(color.FgBlue, color.Bold).SprintFunc()

	newShell := Shell{config: config, db: db, promptFmt: promptFmt}

	dbCmdConfig := &shellcmd.DbCmdConfig{
		Db:                db,
		OutF:              config.OutF,
		ErrF:              config.ErrF,
		SetInterruptShell: func() { newShell.state.interruptReadEvalPrintLoop = true },
		SetMode:           func(mode enums.PrintMode) { newShell.state.printMode = mode },
		GetMode: func() enums.PrintMode {
			return newShell.state.printMode
		},
	}
	newShell.databaseCmd = shellcmd.CreateNewDatabaseRootCmd(dbCmdConfig)

	err := newShell.resetState()
	if err != nil {
		return nil, err
	}

	return &newShell, nil
}

func (sh *Shell) Run() error {
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

func (sh *Shell) resetState() error {
	var err error
	sh.state.readline, err = sh.newReadline()
	if err != nil {
		return err
	}

	sh.state.insideMultilineStatement = false
	sh.state.statementParts = make([]string, 0)

	sh.state.interruptReadEvalPrintLoop = false

	sh.state.printMode = enums.TABLE_MODE

	return nil
}

type shellAutoCompleter struct {
	suggestCompletion func(input string) []string
}

func (sac *shellAutoCompleter) Do(line []rune, pos int) (newLine [][]rune, lengh int) {
	suggestions := sac.suggestCompletion(string(line[0:pos]))

	if suggestions == nil {
		return nil, 0
	}

	runeSuggestions := make([][]rune, 0, len(suggestions))
	for _, suggestion := range suggestions {
		runeSuggestions = append(runeSuggestions, []rune(suggestion))
	}

	return runeSuggestions, pos
}

func (sh *Shell) newReadline() (*readline.Instance, error) {
	historyFile := GetHistoryFileBasedOnMode(sh.db.Uri, sh.config.HistoryMode, sh.config.HistoryName)

	config := &readline.Config{
		Prompt:          sh.promptFmt(promptNewStatement),
		InterruptPrompt: "^C",
		HistoryFile:     historyFile,
		EOFPrompt:       QUIT_COMMAND,
		Stdin:           io.NopCloser(sh.config.InF),
		Stdout:          sh.config.OutF,
		Stderr:          sh.config.ErrF,
	}

	if !sh.config.DisableAutoCompletion {
		autoCompleter := &shellAutoCompleter{suggestCompletion: suggester.SuggestCompletion}
		config.AutoComplete = autoCompleter
	}

	return readline.NewEx(config)
}

func isCommand(line string) bool {
	return line[0] == '.'
}

func (sh *Shell) executeCommand(command string) error {
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

func (sh *Shell) appendStatementPartAndExecuteIfFinished(statementPart string) {
	sh.state.statementParts = append(sh.state.statementParts, statementPart)
	completeStatement := strings.Join(sh.state.statementParts, "\n")
	if isStatementFinished(completeStatement) {
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

func (sh *Shell) ExecuteCommandOrStatements(commandOrStatements string) error {
	if isCommand(commandOrStatements) {
		return sh.executeCommand(commandOrStatements)
	}

	return sh.db.ExecuteAndPrintStatements(commandOrStatements, sh.config.OutF, false, sh.state.printMode)
}

func (sh *Shell) getWelcomeMessage() string {
	if sh.config.WelcomeMessage == nil {
		return DEFAULT_WELCOME_MESSAGE
	}
	return *sh.config.WelcomeMessage
}

func (sh *Shell) CancelQuery() {
	sh.db.CancelQuery()
}

func isStatementFinished(statement string) bool {
	_, splitExtraInfos := sqliteparserutils.SplitStatement(statement)
	return !splitExtraInfos.IncompleteCreateTriggerStatement &&
		!splitExtraInfos.IncompleteMultilineComment &&
		splitExtraInfos.LastTokenType == sqliteparser.SQLiteLexerSCOL
}
