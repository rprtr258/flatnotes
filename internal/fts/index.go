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
	// Term -> Document ID -> Term count in document
	InvIndex map[string]map[string]int
	// all Documents
	Documents map[string]D
	// Term -> Term frequency among all documents
	TermFreq map[string]int
}

func NewIndex[D Document]() *Index[D] {
	return &Index[D]{
		InvIndex:  map[string]map[string]int{},
		Documents: map[string]D{},
		TermFreq:  map[string]int{},
	}
}

// add adds documents to the index.
func (idx Index[D]) Add(doc D) {
	for _, text := range doc.Fields() {
		for _, token := range analyze(text) {
			if _, ok := idx.InvIndex[token.Term]; !ok {
				idx.InvIndex[token.Term] = map[string]int{}
			}
			idx.InvIndex[token.Term][doc.ID()]++
			idx.TermFreq[token.Term]++
		}
	}
	idx.Documents[doc.ID()] = doc
}

func (idx Index[D]) Remove(id string) {
	delete(idx.Documents, id)
	for term, docs := range idx.InvIndex {
		idx.TermFreq[term] -= docs[id]
		delete(docs, id)
	}
}

type Hit[D Document] struct {
	Doc   D
	Score float64
	Terms []Term
	Tags  []string
}

// search queries the index for the given text.
func (idx Index[D]) Search(query string, tags []string) []Hit[D] {
	tagScores := map[string]float64{}
	scores := map[string]float64{}
	docTags := map[string][]string{}

	for id, doc := range idx.Documents {
		tgs := lo.Intersect(strings.Split(doc.Fields()["Tags"], " "), tags)
		if len(tgs) > 0 {
			tagScores[id]++
			docTags[id] = tgs
		}
	}

	queryTokens := analyze(query)
	for _, token := range queryTokens {
		for docID, cnt := range idx.InvIndex[token.Term] {
			scores[docID] += float64(cnt) / float64(idx.TermFreq[token.Term])
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

	allDocIDs := lo.Uniq(append(lo.Keys(scores), lo.Keys(tagScores)...))

	return lo.Map(allDocIDs, func(id string, _ int) Hit[D] {
		return Hit[D]{
			Score: scores[id] + tagScores[id]*2, // TODO: title boost 2
			Doc:   idx.Documents[id],
			Terms: queryTokens,
			Tags:  docTags[id],
		}
	})
}
