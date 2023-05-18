package suggester

import (
	"strings"
	"unicode"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/libsql/sqlite-antlr4-parser/sqliteparser"
)

type tokenStreamInspector struct {
	*antlr.CommonTokenStream

	parser *sqliteparser.SQLiteParser

	lastUserToken                 antlr.Token
	expectedTokensAtLastUserInput *antlr.IntervalSet
	ruleBeforeLastUserInput       int
}

func (mts *tokenStreamInspector) LA(n int) int {
	token := mts.CommonTokenStream.LA(n)

	if n == 1 {
		if nextToken := mts.CommonTokenStream.LA(2); nextToken == antlr.TokenEOF && mts.expectedTokensAtLastUserInput == nil {
			mts.expectedTokensAtLastUserInput = mts.parser.GetExpectedTokens()
			mts.lastUserToken = mts.CommonTokenStream.LT(1)
			mts.ruleBeforeLastUserInput = mts.parser.GetState()
		}
	}
	return token
}

func (mts *tokenStreamInspector) LT(k int) antlr.Token {
	token := mts.CommonTokenStream.LT(k)

	if k == 1 {
		if nextToken := mts.CommonTokenStream.LT(2); nextToken.GetTokenType() == antlr.TokenEOF && mts.expectedTokensAtLastUserInput == nil {
			mts.expectedTokensAtLastUserInput = mts.parser.GetExpectedTokens()
			mts.lastUserToken = mts.CommonTokenStream.LT(1)
			mts.ruleBeforeLastUserInput = mts.parser.GetState()
		}
	}
	return token
}

func (mts *tokenStreamInspector) getLastInputTokenAndExpectedTokens() (lastInputToken antlr.Token, expectedTokens []int) {
	if mts.expectedTokensAtLastUserInput == nil {
		return nil, nil
	}

	expectedTokens = make([]int, 0)
	for _, v := range mts.expectedTokensAtLastUserInput.GetIntervals() {
		for j := v.Start; j < v.Stop; j++ {
			expectedTokens = append(expectedTokens, j)
		}
	}

	return mts.lastUserToken, expectedTokens
}

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

	p.Parse()

	return tokenStream.getLastInputTokenAndExpectedTokens()

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
