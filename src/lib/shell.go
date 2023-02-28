package lib

import (
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
)

const QUIT_COMMAND = ".quit"
const WELCOME_MESSAGE = "Welcome to LibSQL shell!\n\nType \".quit\" to exit the shell, \".tables\" to list all tables, and \".schema\" to show table schemas.\n\n"

type ShellConfig struct {
	InF         io.Reader
	OutF        io.Writer
	ErrF        io.Writer
	HistoryFile string
	QuietMode   bool
}

func NewReadline(config *ShellConfig) (*readline.Instance, error) {
	return readline.NewEx(&readline.Config{
		Prompt:          "→  ",
		InterruptPrompt: "^C",
		HistoryFile:     config.HistoryFile,
		EOFPrompt:       QUIT_COMMAND,
		Stdin:           io.NopCloser(config.InF),
		Stdout:          config.OutF,
		Stderr:          config.ErrF,
	})
}

func (db *Db) RunShell(config *ShellConfig) error {
	l, err := NewReadline(config)
	if err != nil {
		return err
	}
	defer l.Close()
	l.CaptureExitSignal()

	if !config.QuietMode {
		fmt.Print(WELCOME_MESSAGE)
	}

	for {
		line, err := l.Readline()

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
		case line == QUIT_COMMAND:
			return nil
		case isCommand(line):
			db.executeCommand(line, config)
		default:
			db.ExecuteAndPrintStatements(line, l.Stdout(), l.Stderr(), false)
		}

	}
	return nil
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

func (db *Db) executeCommand(command string, config *ShellConfig) {
	statement, isSqlAliasCommands := sqlAliasCommands[command]
	if isSqlAliasCommands {
		db.ExecuteAndPrintStatements(statement, config.OutF, config.ErrF, true)
	} else {
		fmt.Println("Unknown command")
	}
}
