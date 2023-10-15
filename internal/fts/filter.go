package fts

import (
	"strings"

	snowballeng "github.com/kljensen/snowball/english"
	"github.com/samber/lo"
)

// lowercaseFilter returns a slice of tokens normalized to lower case.
func lowercaseFilter(tokens []Term) []Term {
	return lo.Map(tokens, func(term Term, _ int) Term {
		term.Term = strings.ToLower(term.Term)
		return term
	})
}

var stopwords = lo.SliceToMap([]string{
	// TODO: get back when tags are working
	// "a", "and", "be", "have", "i",
	// "in", "of", "that", "the", "to",
}, func(token string) (string, struct{}) {
	return token, struct{}{}
})

// stopwordFilter returns a slice of tokens with stop words removed.
func stopwordFilter(tokens []Term) []Term {
	return lo.Filter(tokens, func(term Term, _ int) bool {
		_, ok := stopwords[term.Term]
		return !ok
	})
}

// stemmerFilter returns a slice of stemmed tokens.
func stemmerFilter(tokens []Term) []Term {
	return lo.Map(tokens, func(token Term, _ int) Term {
		token.Term = snowballeng.Stem(token.Term, false)
		return token
	})
}
