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
	InF  io.Reader
	OutF io.Writer
	ErrF io.Writer
}

func NewReadline(config *ShellConfig) (*readline.Instance, error) {
	return readline.NewEx(&readline.Config{
		Prompt:          "→  ",
		InterruptPrompt: "^C",
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

	fmt.Print(WELCOME_MESSAGE)

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
			db.ExecuteAndPrintStatements(line, l.Stdout(), l.Stderr())
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
		order by name`,
	".schema": `select sql from sqlite_schema
		where name not like 'sqlite_%'
		order by name`,
}

func (db *Db) executeCommand(command string, config *ShellConfig) {
	statement, isSqlAliasCommands := sqlAliasCommands[command]
	if isSqlAliasCommands {
		db.options.withoutHeader = true
		db.ExecuteAndPrintStatements(statement, config.OutF, config.ErrF)
		db.options.withoutHeader = false
	} else {
		fmt.Println("Unknown command")
	}
}
