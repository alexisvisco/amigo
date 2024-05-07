package utils

import (
	"fmt"
	"strings"
	"unicode"
)

// RemoveConsecutiveSpaces Remove all whitespace not between matching unescaped quotes.
// Example: SELECT    * FROM "table" WHERE "column" = 'value  '
// Result: SELECT * FROM "table" WHERE "column" = 'value  '
func RemoveConsecutiveSpaces(s string) (string, error) {
	rs := make([]rune, 0, len(s))
	for i := 0; i < len(s); i++ {
		r := rune(s[i])
		if r == '\'' || r == '"' {
			prevChar := ' '
			matchedChar := uint8(r)

			// if the text remaining is 'value \' '
			// then the quoteText will be 'value \' '
			// if there is no end quote then it will return an error
			quoteText := string(s[i])

			// jump until the next matching quote character
			for n := i + 1; n < len(s); n++ {
				if s[n] == matchedChar && prevChar != '\\' {
					i = n
					quoteText += string(s[n])
					break
				}
				quoteText += string(s[n])
				prevChar = rune(s[n])
			}

			if quoteText[len(quoteText)-1] != matchedChar || len(quoteText) == 1 {
				err := fmt.Errorf("unmatched unescaped quote: %q", quoteText)
				return "", err
			}

			rs = append(rs, []rune(quoteText)...)
			continue
		}

		if unicode.IsSpace(r) {
			rs = append(rs, r)

			// jump until the next non-space character
			for n := i + 1; n < len(s); n++ {
				if !unicode.IsSpace(rune(s[n])) {
					i = n - 1 // -1 because the loop will increment it
					break
				}
			}

			continue
		}

		if !unicode.IsSpace(r) {
			rs = append(rs, r)
		}

	}

	return strings.Trim(string(rs), " "), nil
}

func Parentheses(in string) string { return fmt.Sprintf("(%s)", in) }

type Replacer map[string]func() string

// Replace replaces the string with the given values {<key>} to the value of the function
func (r Replacer) Replace(str string) string {
	for k, v := range r {
		str = strings.ReplaceAll(str, "{"+k+"}", v())
	}
	res, err := RemoveConsecutiveSpaces(str)
	if err != nil {
		res = str
	}

	return res
}

// StrFunc returns a function that returns the given value as a string
func StrFunc[T any](val T) func() string { return func() string { return fmt.Sprint(val) } }

func StrFuncPredicate[T any](condition bool, val T) func() string {
	return func() string {
		if condition {
			return fmt.Sprint(val)
		}
		return ""
	}
}
