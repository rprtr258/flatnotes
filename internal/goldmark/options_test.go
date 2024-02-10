package goldmark_test

import (
	"testing"

	. "github.com/rprtr258/flatnotes/internal/goldmark"
	"github.com/rprtr258/flatnotes/internal/goldmark/parser"
	"github.com/rprtr258/flatnotes/internal/goldmark/testutil"
)

func TestAttributeAndAutoHeadingID(t *testing.T) {
	markdown := New(
		WithParserOptions(
			parser.WithAttribute(),
			parser.WithAutoHeadingID(),
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/options.txt", t, testutil.ParseCliCaseArg()...)
}
