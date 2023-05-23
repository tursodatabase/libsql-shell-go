package suggester

import (
	"strings"
	"unicode"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/libsql/sqlite-antlr4-parser/sqliteparser"
)

func SuggestCompletion(currentInput string) []string {
	suggestions := make([]string, 0)

	lastInputToken, expectedTokens := getLastInputTokenAndExpectedTokens(currentInput)
	if expectedTokens == nil {
		return nil
	}

	suggestionsShouldBeLowerCase := calculateSuggestionLetterCaseType(currentInput)

	expectedLiteralTokenStrings := getLiteralTokenStrings(expectedTokens)

	for _, expectedLiteralTokenString := range expectedLiteralTokenStrings {
		uppercasedLastInputToken := strings.ToUpper(lastInputToken.GetText())

		if suggestion, isPrefix := strings.CutPrefix(expectedLiteralTokenString, uppercasedLastInputToken); isPrefix {
			if suggestionsShouldBeLowerCase {
				suggestion = strings.ToLower(suggestion)
			} else {
				suggestion = strings.ToUpper(suggestion)
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions
}

func getLastInputTokenAndExpectedTokens(currentInput string) (lastInputToken antlr.Token, expectedTokens []int) {

	currentInputStream := antlr.NewInputStream(currentInput)

	lexer := sqliteparser.NewSQLiteLexer(currentInputStream)
	tokenStream := &tokenStreamInspector{CommonTokenStream: antlr.NewCommonTokenStream(lexer, 0)}

	p := sqliteparser.NewSQLiteParser(tokenStream)
	p.RemoveErrorListeners()
	p.SetErrorHandler(antlr.NewBailErrorStrategy())

	tokenStream.parser = p

	defer func() {
		if r := recover(); r != nil {
			lastInputToken, expectedTokens = tokenStream.getLastInputTokenAndExpectedTokens()
		}
	}()

	tree := p.Parse()

	lastInputToken, expectedTokens = tokenStream.getLastInputTokenAndExpectedTokens()

	lastInputTokenRules := getTokenRules(tree, lastInputToken)

	expectedTokens = filterExpectedTokensBasedOnTokenRules(lastInputTokenRules, expectedTokens)

	return lastInputToken, expectedTokens
}

var expectedTokensFilterByRule = map[int]func(expectedTokens []int) []int{
	sqliteparser.SQLiteParserRULE_any_name: noExpectedToken,
}

func filterExpectedTokensBasedOnTokenRules(tokenRules []int, expectedTokens []int) []int {
	for _, tokenRule := range tokenRules {
		if filterFunc, ok := expectedTokensFilterByRule[tokenRule]; ok {
			return filterFunc(expectedTokens)
		}
	}

	return expectedTokens
}

func noExpectedToken(_ []int) []int {
	return make([]int, 0)
}

func getLiteralTokenStrings(tokens []int) []string {
	literalTokenNames, _ := getLiteralAndSymbolicNames()

	tokenStrings := make([]string, 0)
	for _, token := range tokens {
		if isSQLiteSpecificToken(token) && token < len(literalTokenNames) && literalTokenNames[token] != "" {
			tokenStrings = append(tokenStrings, strings.Trim(literalTokenNames[token], "'"))
		}
	}

	return tokenStrings
}

func getLiteralAndSymbolicNames() ([]string, []string) {
	mockStream := antlr.NewInputStream("")
	mockLexer := sqliteparser.NewSQLiteLexer(mockStream)

	return mockLexer.LiteralNames, mockLexer.SymbolicNames
}

func isSQLiteSpecificToken(token int) bool {
	return token > 0
}

func calculateSuggestionLetterCaseType(currentInput string) (shouldBeLowerCase bool) {
	for i := len(currentInput) - 1; i >= 0; i-- {
		r := rune(currentInput[i])
		if !unicode.IsLetter(r) {
			continue
		}

		if unicode.IsLower(r) {
			return true
		} else if unicode.IsUpper(r) {
			return false
		}
	}

	return false
}
