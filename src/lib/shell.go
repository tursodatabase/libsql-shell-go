package lib

import (
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/c-bata/go-prompt"
)

const QUIT_COMMAND = ".quit"
const WELCOME_MESSAGE = "Welcome to LibSQL shell!\n\nType \".quit\" to exit the shell.\n\n"

func NewPrompt(in io.Reader, out io.Writer, err io.Writer, db *sql.DB) *prompt.Prompt {
	return prompt.New(
		(func(in string) {
			executor(in, db)
		}),
		completer,
		prompt.OptionPrefix("→ "),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,

			Fn: func(buf *prompt.Buffer) {
				fmt.Println(io.EOF)
			},
		}),
	)
}

func RunShell(p *prompt.Prompt) {
	p.Run()
}

func executor(in string, db *sql.DB) {
	line := strings.TrimSpace(in)
	if len(line) == 0 {
		return
	}
	switch line {
	case QUIT_COMMAND:
		// TODO: Exit the shell
		break
	default:
		result, err := ExecuteStatements(db, line)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		} else {
			fmt.Println(result)
		}
	}
}

func completer(in prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}
