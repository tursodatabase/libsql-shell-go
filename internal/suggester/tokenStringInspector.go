package suggester

import (
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
