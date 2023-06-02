package suggester_test

import (
	"testing"

	"github.com/libsql/libsql-shell-go/internal/suggester"
	"github.com/stretchr/testify/assert"
)

func Test_GivenIncompleteKeywordInput_WhenSuggestCompletion_ExpectValidKeywords(t *testing.T) {
	tests := []struct {
		input              string
		expectedSuggestion []string
	}{
		{
			input:              "s",
			expectedSuggestion: []string{"avepoint", "elect"},
		},
		{
			input:              "se",
			expectedSuggestion: []string{"lect"},
		},
		{
			input:              "select ",
			expectedSuggestion: []string{},
		},
		{
			input:              "select 1 ",
			expectedSuggestion: []string{},
		},
		{
			input:              "select 1 f",
			expectedSuggestion: []string{"rom"},
		},
		{
			input:              "select 1 from tableName; se",
			expectedSuggestion: []string{"lect"},
		},
		// TODO: Fix this test
		// {
		// 	input:              "select 1 from tableName; select 1 f",
		// 	expectedSuggestion: []string{"rom"},
		// },
		{
			input:              "c",
			expectedSuggestion: []string{"ommit", "reate"},
		},
		{
			input:              "i",
			expectedSuggestion: []string{"nsert"},
		},
		// TODO: Fix this test
		// {
		// 	input:              "insert i",
		// 	expectedSuggestion: []string{"nto"},
		// },
		{
			input:              "inse ",
			expectedSuggestion: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			assert := assert.New(t)

			gotSuggestion := suggester.SuggestCompletion(test.input)

			assert.ElementsMatch(test.expectedSuggestion, gotSuggestion)
		})
	}
}

func Test_GivenIncompleteTableName_WhenSuggestCompletion_ExpectNoSuggestion(t *testing.T) {
	assert := assert.New(t)

	gotSuggestion := suggester.SuggestCompletion("select 1 from t")

	assert.ElementsMatch([]string{}, gotSuggestion)
}
