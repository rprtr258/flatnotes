package fts

import (
	"unicode"

	snowballeng "github.com/kljensen/snowball/english"
	"github.com/rprtr258/fun/iter"
)

type Term struct {
	Term string
	I, J int
}

// tokenize returns a slice of tokens for the given text.
func tokenize(s string) iter.Seq[Term] {
	return func(yield func(Term) bool) bool {
		// Find the field start and end indices.
		// Doing this in a separate pass (rather than slicing the string s
		// and collecting the result substrings right away) is significantly
		// more efficient, possibly due to cache effects.
		start := -1 // valid span start if >= 0
		for end, r := range s {
			if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
				if start >= 0 {
					if !yield(Term{
						Term: s[start:end],
						I:    start,
						J:    end,
					}) {
						return false
					}
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
			if !yield(Term{
				Term: s[start:],
				I:    start,
				J:    len(s),
			}) {
				return false
			}
		}
		return true
	}
}

// analyze analyzes the text and returns a slice of tokens.
func analyze(text string) iter.Seq[Term] {
	return tokenize(text).
		Map(func(term Term) Term {
			// stemmed tokens
			term.Term = snowballeng.Stem(term.Term, false)
			return term
		})
}
