package highlights

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// NormaliseSpaces turns all whitespace characters into the same space
// and strips out any multiples
func NormaliseSpaces(s string) string {
	ct := 0

	return strings.Map(func(r rune) rune {
		if !unicode.IsSpace(r) {
			ct = 0
			return r
		}

		ct++
		if ct > 1 {
			return -1
		}

		return ' '
	}, s)
}

// NormaliseString runs all the normalisation functions used
// currently, it does the following
// - converts to lowercase
// - Normalises all equivalent characters using unicode NFKC
// - Strips successive spaces and converts all space characters to be the same
func NormaliseString(s string) string {
	for _, fn := range normalisationFuncs {
		s = fn(s)
	}
	return s
}

var normalisationFuncs = []func(string) string{norm.NFKC.String, strings.ToLower, NormaliseSpaces}
