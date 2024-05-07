package utils

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRemoveConsecutiveSpace(t *testing.T) {
	testCases := []struct {
		input, expectedOutput string
		err                   require.ErrorAssertionFunc
	}{
		{
			input:          `SELECT    * FROM "table" WHERE "column" = 'value  '`,
			expectedOutput: `SELECT * FROM "table" WHERE "column" = 'value  '`,
		},
		{
			input:          `Hello  "world  "`,
			expectedOutput: `Hello "world  "`,
		},
		{
			input:          `a     b     c`,
			expectedOutput: `a b c`,
		},
		{
			input:          `"a b"    "c d"`,
			expectedOutput: `"a b" "c d"`,
		},
		{
			input: `abc 'a`,
			err:   require.Error,
		},
	}

	for _, testCase := range testCases {
		result, err := RemoveConsecutiveSpaces(testCase.input)

		if testCase.err != nil {
			testCase.err(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, testCase.expectedOutput, result)
		}

	}
}
