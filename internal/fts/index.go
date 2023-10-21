package fts

import (
	"strings"

	"github.com/samber/lo"
)

type Document interface {
	ID() string
	Fields() map[string]string
}

// Index is an inverted Index. It maps tokens to document IDs.
type Index[D Document] struct {
	// Term -> Document IDs
	InvIndex map[string][]string
	// all Documents
	Documents map[string]D
}

func NewIndex[D Document]() *Index[D] {
	return &Index[D]{
		InvIndex:  map[string][]string{},
		Documents: map[string]D{},
	}
}

// add adds documents to the index.
func (idx Index[D]) Add(doc D) {
	for _, text := range doc.Fields() {
		for _, token := range analyze(text) {
			if lo.Contains(idx.InvIndex[token.Term], doc.ID()) {
				// Don't add same ID twice.
				continue
			}
			idx.InvIndex[token.Term] = append(idx.InvIndex[token.Term], doc.ID())
		}
	}
	idx.Documents[doc.ID()] = doc
}

// add adds documents to the index.
func (idx Index[D]) Remove(id string) {
	delete(idx.Documents, id)
	for token, ids := range idx.InvIndex {
		idx.InvIndex[token] = lo.Filter(
			ids,
			func(docID string, _ int) bool {
				return docID == id
			})
	}
}

// intersection returns the set intersection between a and b.
// a and b have to be sorted in ascending order and contain no duplicates.
func intersection(a, b []string) []string {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	r := make([]string, 0, maxLen)
	var i, j int
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			i++
		} else if a[i] > b[j] {
			j++
		} else {
			r = append(r, a[i])
			i++
			j++
		}
	}
	return r
}

type Hit[D Document] struct {
	Doc   D
	Score float64
	Terms []Term
}

// search queries the index for the given text.
func (idx Index[D]) Search(query string, tags []string) []Hit[D] {
	tagDocIDs := map[string]float64{}

	for id, doc := range idx.Documents {
		if len(lo.Intersect(strings.Split(doc.Fields()["Tags"], " "), tags)) > 0 {
			tagDocIDs[id]++
		}
	}

	var docIDs []string
	queryTokens := analyze(query)
	for _, token := range queryTokens {
		if ids, ok := idx.InvIndex[token.Term]; ok {
			if docIDs == nil {
				docIDs = ids
			} else {
				docIDs = intersection(docIDs, ids)
			}
		} else {
			// Token doesn't exist.
			return nil
		}
		// for term, ids := range idx.InvIndex { // stupidest fuzzy search in the world starts here
		// 	if !strings.Contains(term, token.Term) {
		// 		continue
		// 	}
		// 	if r == nil {
		// 		r = ids
		// 	} else {
		// 		r = intersection(r, ids)
		// 	}
		// }
	}
	for _, id := range docIDs {
		tagDocIDs[id]++
	}

	return lo.MapToSlice(tagDocIDs, func(id string, score float64) Hit[D] {
		return Hit[D]{
			Score: score,
			Doc:   idx.Documents[id],
			Terms: queryTokens,
		}
	})
}
