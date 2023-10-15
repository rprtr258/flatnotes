package fts

import "unicode"

type Term struct {
	Term string
	I, J int
}

// tokenize returns a slice of tokens for the given text.
func tokenize(s string) []Term {
	// A span is used to record a slice of s of the form s[start:end].
	// The start index is inclusive and the end index is exclusive.
	type span struct {
		start int
		end   int
	}
	spans := make([]span, 0, 32)

	// Find the field start and end indices.
	// Doing this in a separate pass (rather than slicing the string s
	// and collecting the result substrings right away) is significantly
	// more efficient, possibly due to cache effects.
	start := -1 // valid span start if >= 0
	for end, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			if start >= 0 {
				spans = append(spans, span{start, end})
				// Set start to a negative value.
				// Note: using -1 here consistently and reproducibly
				// slows down this code by a several percent on amd64.
				start = ^start
			}
		} else {
			if start < 0 {
				start = end
			}
		}
	}

	// Last field might end at EOF.
	if start >= 0 {
		spans = append(spans, span{start, len(s)})
	}

	// Create strings from recorded field indices.
	a := make([]Term, len(spans))
	for i, span := range spans {
		a[i] = Term{
			Term: s[span.start:span.end],
		}
	}
	return a
}

// analyze analyzes the text and returns a slice of tokens.
func analyze(text string) []Term {
	tokens := tokenize(text)
	tokens = lowercaseFilter(tokens)
	tokens = stopwordFilter(tokens)
	tokens = stemmerFilter(tokens)
	return tokens
}
