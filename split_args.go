package exec

import (
	"strings"
)

const (
	quoteNone   = iota
	quoteDouble = 1
	quoteSingle = 2
)

func SplitArgs(s string) []string {
	quote := quoteNone
	token := strings.Builder{}
	tokens := []string{}
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		c := runes[i]

		if quote != quoteNone {
			switch quote {
			case quoteSingle:
				if c == '\'' {
					quote = quoteNone
					if token.Len() > 0 {
						tokens = append(tokens, token.String())
						token.Reset()
					}

					continue
				}
			case quoteDouble:
				if c == '"' {
					quote = quoteNone
					if token.Len() > 0 {
						tokens = append(tokens, token.String())
						token.Reset()
					}

					continue
				}
			}

			token.WriteRune(c)
			continue
		}

		if c == ' ' {
			remaining := len(runes) - 1 - i
			if remaining > 2 {
				j := runes[i+1]
				k := runes[i+2]

				if j == '\n' {
					i += 1
					if token.Len() > 0 {
						tokens = append(tokens, token.String())
						token.Reset()
					}

					continue
				}

				if j == '\r' && k == '\n' {
					i += 2
					if token.Len() > 0 {
						tokens = append(tokens, token.String())
						token.Reset()
					}

					continue
				}

				if (j == '\\' || j == '`') && k == '\n' {
					i += 2

					if token.Len() > 0 {
						tokens = append(tokens, token.String())
					}

					token.Reset()
					continue
				}

				if remaining > 3 {
					l := runes[i+3]
					if (j == '\\' || j == '`') && k == '\r' && l == '\n' {
						i += 3
						if token.Len() > 0 {
							tokens = append(tokens, token.String())
						}

						token.Reset()
						continue
					}
				}

				if token.Len() > 0 {
					tokens = append(tokens, token.String())
					token.Reset()
				}

				continue
			}
		}

		if token.Len() == 0 {
			switch c {
			case '\'':
				quote = quoteSingle
				continue

			case '"':
				quote = quoteDouble
				continue
			}
		}

		token.WriteRune(c)
	}

	if token.Len() > 0 {
		tokens = append(tokens, token.String())
	}

	token.Reset()

	return tokens
}
