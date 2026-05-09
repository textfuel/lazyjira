// Package ascii reduces Unicode strings to ASCII via transliteration
// and NFD-based accent stripping.
package ascii

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

// transliterations expands letters that need multi-char ASCII output
// (German ä->ae, ß->ss). Plain accents go through NFD in Convert.
var transliterations = map[rune]string{
	'ä': "ae",
	'ö': "oe",
	'ü': "ue",
	'ß': "ss",
}

// Convert lowercases s and reduces it to ASCII. German umlauts and ß
// expand (ä->ae, ß->ss); other accented letters decompose via NFD to
// their base letter (é->e). Anything still non-ASCII is dropped. ASCII
// punctuation is preserved - callers decide how to filter it.
func Convert(s string) string {
	s = strings.ToLower(s)

	var pre strings.Builder
	pre.Grow(len(s))
	for _, r := range s {
		if rep, ok := transliterations[r]; ok {
			pre.WriteString(rep)
		} else {
			pre.WriteRune(r)
		}
	}
	decomposed := norm.NFD.String(pre.String())

	var out strings.Builder
	out.Grow(len(decomposed))
	for _, r := range decomposed {
		if r >= 128 {
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}
