package extension

import (
	"testing"

	"github.com/rprtr258/flatnotes/internal/goldmark"
	"github.com/rprtr258/flatnotes/internal/goldmark/renderer/html"
	"github.com/rprtr258/flatnotes/internal/goldmark/testutil"
)

func TestTypographer(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			NewTypographer(),
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/typographer.txt", t, testutil.ParseCliCaseArg()...)
}
