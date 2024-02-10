package fts

import (
	"sync"

	"github.com/rprtr258/fun"
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
	mu sync.RWMutex
	// Field -> Term -> Document ID -> Term count in document field
	InvIndex map[string]map[string]map[string]int
	// Field -> Term -> Term frequency among all documents field
	TermFreq map[string]map[string]int
}

func NewIndex[D Document]() *Index[D] {
	InvIndex := map[string]map[string]map[string]int{}
	TermFreq := map[string]map[string]int{}
	for field := range fun.Zero[D]().Fields() {
		if _, ok := InvIndex[field]; !ok {
			InvIndex[field] = map[string]map[string]int{}
		}

		if _, ok := TermFreq[field]; !ok {
			TermFreq[field] = map[string]int{}
		}
	}

	return &Index[D]{
		mu:       sync.RWMutex{},
		InvIndex: InvIndex,
		TermFreq: TermFreq,
	}
}

func (idx *Index[D]) add(field, term, docID string, cnt int) {
	if _, ok := idx.InvIndex[field][term]; !ok {
		idx.InvIndex[field][term] = map[string]int{}
	}

	idx.InvIndex[field][term][docID] += cnt
	idx.TermFreq[field][term] += cnt
}

// add adds documents to the index.
func (idx *Index[D]) Add(docs ...D) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for _, doc := range docs {
		for fieldName, field := range doc.Fields() {
			analyze(field.Content)(func(term Term) bool {
				idx.add(fieldName, term.Term, doc.ID(), 1)
				return true
			})
			for _, term := range field.Terms {
				idx.add(fieldName, term, doc.ID(), 1)
			}
		}
	}
}

func (idx *Index[D]) Delete(id string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for field, index := range idx.InvIndex {
		for term, docs := range index {
			idx.TermFreq[field][term] -= docs[id]
			delete(docs, id)
		}
	}
}

type Hit[D Document] struct {
	ID    string
	Score float64
	Terms []Term
	Tags  []string
}

// search queries the index for the given text.
func (idx *Index[D]) Search(query string, tags []string) []Hit[D] {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	scores := map[string]float64{}
	docTags := map[string][]string{}

	queryTokens := analyze(query).ToSlice()
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

	return fun.MapToSlice(scores, func(id string, score float64) Hit[D] {
		return Hit[D]{
			ID:    id,
			Score: score,
			Terms: queryTokens,
			Tags:  docTags[id],
		}
	})
}
