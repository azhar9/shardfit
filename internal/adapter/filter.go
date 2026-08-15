package adapter

import "strings"

// SplitFilter tokenizes a filter string into argv elements, honoring single
// and double quotes so runner-native expressions like `-k "create or
// update"` survive as one token. shardfit never interprets filter semantics
// — it only splits the string on the user's behalf. An unmatched quote is
// dropped and the remainder becomes one token. An empty quoted token is
// dropped too (`-k ""` → `["-k"]`), so runners whose empty argument means
// match-all need the option omitted rather than empty-quoted.
func SplitFilter(s string) []string {
	out := []string{}
	var cur strings.Builder
	var quote byte
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			} else {
				cur.WriteByte(c)
			}
		case c == '"' || c == '\'':
			quote = c
		case c == ' ' || c == '\t':
			flush()
		default:
			cur.WriteByte(c)
		}
	}
	flush()
	return out
}
