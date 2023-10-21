package fts

import (
	"strings"

	snowballeng "github.com/kljensen/snowball/english"
	"github.com/rprtr258/fun/iter"
)

// lowercaseFilter returns a slice of tokens normalized to lower case.
func lowercaseFilter(terms iter.Seq[Term]) iter.Seq[Term] {
	return terms.Map(func(term Term) Term {
		term.Term = strings.ToLower(term.Term)
		return term
	})
}

// stemmerFilter returns a slice of stemmed tokens.
func stemmerFilter(terms iter.Seq[Term]) iter.Seq[Term] {
	return terms.Map(func(term Term) Term {
		term.Term = snowballeng.Stem(term.Term, false)
		return term
	})
}
