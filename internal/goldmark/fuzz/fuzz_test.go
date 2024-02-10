package fuzz

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/rprtr258/flatnotes/internal/goldmark"
	"github.com/rprtr258/flatnotes/internal/goldmark/extension"
	"github.com/rprtr258/flatnotes/internal/goldmark/parser"
	"github.com/rprtr258/flatnotes/internal/goldmark/renderer/html"
)

func fuzz(f *testing.F) {
	f.Fuzz(func(t *testing.T, orig string) {
		markdown := goldmark.New(
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
				parser.WithAttribute(),
			),
			goldmark.WithRendererOptions(
				html.WithUnsafe(),
				html.WithXHTML(),
			),
			goldmark.WithExtensions(
				extension.DefinitionList,
				extension.NewFootnote(),
				extension.GFM,
				extension.NewTypographer(),
				extension.Linkify,
				extension.Table,
				extension.TaskList,
			),
		)
		var b bytes.Buffer
		if err := markdown.Convert([]byte(orig), &b); err != nil {
			panic(err)
		}
	})
}

func FuzzDefault(f *testing.F) {
	bs, err := os.ReadFile("../_test/spec.json")
	if err != nil {
		panic(err)
	}
	var testCases []map[string]any
	if err := json.Unmarshal(bs, &testCases); err != nil {
		panic(err)
	}
	for _, c := range testCases {
		f.Add(c["markdown"])
	}
	fuzz(f)
}

func FuzzOss(f *testing.F) {
	fuzz(f)
}
