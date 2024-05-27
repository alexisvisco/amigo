package pg

import (
	"strings"
)

func QuoteIdent(name string) string {
	end := strings.IndexRune(name, 0)
	if end > -1 {
		name = name[:end]
	}
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func QuoteValue(v interface{}) string {
	switch v := v.(type) {
	default:
		panic("unsupported value")
	case string:
		v = strings.ReplaceAll(v, `'`, `''`)
		if strings.Contains(v, `\`) {
			v = strings.ReplaceAll(v, `\`, `\\`)
			v = ` E'` + v + `'`
		} else {
			v = `'` + v + `'`
		}
		return v
	}
}
