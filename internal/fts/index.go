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
	// Field -> Term -> Document ID -> Term count in document field
	InvIndex map[string]map[string]map[string]int
	// all Documents
	Documents map[string]D
	// Field -> Term -> Term frequency among all documents field
	TermFreq map[string]map[string]int
}

func NewIndex[D Document]() *Index[D] {
	InvIndex := map[string]map[string]map[string]int{}
	TermFreq := map[string]map[string]int{}
	for field := range func() D {
		var d D
		return d
	}().Fields() {
		if _, ok := InvIndex[field]; !ok {
			InvIndex[field] = map[string]map[string]int{}
		}

		if _, ok := TermFreq[field]; !ok {
			TermFreq[field] = map[string]int{}
		}
	}

	return &Index[D]{
		InvIndex:  InvIndex,
		Documents: map[string]D{},
		TermFreq:  TermFreq,
	}
}

func (idx Index[D]) add(field, term, docID string, cnt int) {
	if _, ok := idx.InvIndex[field][term]; !ok {
		idx.InvIndex[field][term] = map[string]int{}
	}

	idx.InvIndex[field][term][docID] += cnt

	idx.TermFreq[field][term] += cnt
}

// add adds documents to the index.
func (idx Index[D]) Add(doc D) {
	for fieldName, field := range doc.Fields() {
		for _, token := range analyze(field.Content) {
			idx.add(fieldName, token.Term, doc.ID(), 1)
		}
		for _, term := range field.Terms {
			idx.add(fieldName, term, doc.ID(), 1)
		}
	}
	idx.Documents[doc.ID()] = doc
}

func (idx Index[D]) Remove(id string) {
	for field := range idx.Documents[id].Fields() {
		for term, docs := range idx.InvIndex[field] {
			idx.TermFreq[field][term] -= docs[id]
			delete(docs, id)
		}
	}
	delete(idx.Documents, id)
}

type Hit[D Document] struct {
	Doc   D
	Score float64
	Terms []Term
	Tags  []string
}

// search queries the index for the given text.
func (idx Index[D]) Search(query string, tags []string) []Hit[D] {
	scores := map[string]float64{}
	docTags := map[string][]string{}

	queryTokens := analyze(query)
	for fieldName, field := range func() D {
		var d D
		return d
	}().Fields() {
		for _, token := range queryTokens {
			for docID, cnt := range idx.InvIndex[fieldName][token.Term] {
				if idx.TermFreq[fieldName][token.Term] == 0 {
					continue
				}

				scores[docID] += float64(cnt) / float64(idx.TermFreq[fieldName][token.Term]) * field.Weight
			}
		}
	}

	return lo.MapToSlice(scores, func(id string, score float64) Hit[D] {
		return Hit[D]{
			Score: score,
			Doc:   idx.Documents[id],
			Terms: queryTokens,
			Tags:  docTags[id],
		}
	})
}
