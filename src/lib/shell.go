package lib

import (
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
)

const QUIT_COMMAND = ".quit"
const WELCOME_MESSAGE = "Welcome to LibSQL shell!\n\nType \".quit\" to exit the shell, \".tables\" to list all tables, and \".schema\" to show table schemas.\n\n"

const promptNewStatement = "→  "
const promptContinueStatement = "... "

type ShellConfig struct {
	InF         io.Reader
	OutF        io.Writer
	ErrF        io.Writer
	HistoryFile string
	QuietMode   bool
}

type shell struct {
	config ShellConfig

	db *Db

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
	return &shell{config: config, db: db}, nil
}

func (sh *shell) run() error {
	var err error
	sh.readline, err = newReadline(&sh.config)
	if err != nil {
		return err
	}
	defer sh.readline.Close()
	sh.readline.CaptureExitSignal()

	if !sh.config.QuietMode {
		fmt.Print(WELCOME_MESSAGE)
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
			sh.executeCommand(line)
		default:
			sh.appendStatementPartAndExecuteIfFinished(line)
		}

	}
	return nil
}

func newReadline(config *ShellConfig) (*readline.Instance, error) {
	return readline.NewEx(&readline.Config{
		Prompt:          promptNewStatement,
		InterruptPrompt: "^C",
		HistoryFile:     config.HistoryFile,
		EOFPrompt:       QUIT_COMMAND,
		Stdin:           io.NopCloser(config.InF),
		Stdout:          config.OutF,
		Stderr:          config.ErrF,
	})
}

func isCommand(line string) bool {
	return line[0] == '.'
}

var sqlAliasCommands = map[string]string{
	".tables": `select name from sqlite_schema
		where type = 'table'
		and name not like 'sqlite_%'
		and name != '_litestream_seq'
		and name != '_litestream_lock'
		and name != 'libsql_wasm_func_table'
		order by name`,
	".schema": `select sql from sqlite_schema
		where name not like 'sqlite_%'
		and name != '_litestream_seq'
		and name != '_litestream_lock'
		and name != 'libsql_wasm_func_table'
		order by name`,
}

func (sh *shell) executeCommand(command string) {
	statement, isSqlAliasCommands := sqlAliasCommands[command]
	if isSqlAliasCommands {
		sh.db.ExecuteAndPrintStatements(statement, sh.config.OutF, sh.config.ErrF, true)
	} else {
		fmt.Println("Unknown command")
	}
}

func (sh *shell) appendStatementPartAndExecuteIfFinished(statementPart string) {
	sh.statementParts = append(sh.statementParts, statementPart)
	if strings.HasSuffix(statementPart, ";") {
		completeStatement := strings.Join(sh.statementParts, "\n")
		sh.statementParts = make([]string, 0)
		sh.insideMultilineStatement = false
		sh.readline.SetPrompt(promptNewStatement)
		sh.db.ExecuteAndPrintStatements(completeStatement, sh.readline.Stdout(), sh.readline.Stderr(), false)
	} else {
		sh.readline.SetPrompt(promptContinueStatement)
		sh.insideMultilineStatement = false
	}
}
