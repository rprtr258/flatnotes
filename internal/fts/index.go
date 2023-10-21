package fts

import (
	"github.com/samber/lo"
)

type DocumentField struct {
	Content string
	Terms   []string
	Weight  float64
}

type Document interface {
	ID() string
	Fields() map[string]DocumentField
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
		for _, token := range analyze(text.Content) {
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

	queryTokens := analyze(query)
	for _, token := range queryTokens {
		for docID, cnt := range idx.InvIndex[token.Term] {
			scores[docID] += float64(cnt) / float64(idx.TermFreq[token.Term])
			// for id, doc := range idx.Documents {
			// 	tagsField := doc.Fields()["Tags"]
			// 	tgs := lo.Intersect(tagsField.Terms, tags)
			// 	tagScores[id] += float64(len(tgs)) * tagsField.Weight
			// 	docTags[id] = tgs
			// }
		}
	}

	allDocIDs := lo.Uniq(append(lo.Keys(scores), lo.Keys(tagScores)...))

	return lo.Map(allDocIDs, func(id string, _ int) Hit[D] {
		return Hit[D]{
			Score: scores[id] + tagScores[id], // TODO: title boost 2
			Doc:   idx.Documents[id],
			Terms: queryTokens,
			Tags:  docTags[id],
		}
	})
}
