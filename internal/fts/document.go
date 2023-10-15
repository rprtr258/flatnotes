package fts

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/samber/lo"
)

// document represents a Wikipedia abstract dump document.
type document struct {
	Id    string
	Title string `xml:"title"`
	URL   string `xml:"url"`
	Text  string `xml:"abstract"`
}

func (d document) ID() string {
	return d.Id
}

func (d document) Fields() map[string]string {
	return map[string]string{
		"Title": d.Title,
		"URL":   d.URL,
		"Text":  d.Text,
	}
}

// loadDocuments loads a Wikipedia abstract dump and returns a slice of documents.
// Dump example: https://dumps.wikimedia.org/enwiki/latest/enwiki-latest-abstract1.xml.gz
func loadDocuments(path string) (map[string]document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	dec := xml.NewDecoder(gz)
	dump := struct {
		Documents []document `xml:"doc"`
	}{}
	if err := dec.Decode(&dump); err != nil {
		return nil, err
	}

	return lo.SliceToMap(dump.Documents, func(doc document) (string, document) {
		return fmt.Sprint(doc), doc
	}), nil
}
