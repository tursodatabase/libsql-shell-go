package suggester

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/libsql/sqlite-antlr4-parser/sqliteparser"
)

type TokenRulesFinder struct {
	sqliteparser.BaseSQLiteParserListener

	targetToken antlr.Token

	tokenRules []int
}

func NewTokenRulesFinder(targetToken antlr.Token) *TokenRulesFinder {
	return &TokenRulesFinder{targetToken: targetToken, tokenRules: make([]int, 0)}
}

func (trf *TokenRulesFinder) EnterEveryRule(ctx antlr.ParserRuleContext) {
	ruleStartTokenIndex := ctx.GetStart().GetTokenIndex()
	ruleEndTokenIndex := ctx.GetStop().GetTokenIndex()
	targetTokenIndex := trf.targetToken.GetTokenIndex()

	if ruleStartTokenIndex <= targetTokenIndex && targetTokenIndex <= ruleEndTokenIndex {
		trf.tokenRules = append(trf.tokenRules, ctx.GetRuleIndex())
	}
}

func getTokenRules(tree sqliteparser.IParseContext, targetToken antlr.Token) []int {

	tokenRulesFinder := NewTokenRulesFinder(targetToken)

	antlr.ParseTreeWalkerDefault.Walk(tokenRulesFinder, tree)

	return tokenRulesFinder.tokenRules
}
