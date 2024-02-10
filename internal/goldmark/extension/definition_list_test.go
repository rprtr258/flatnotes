package extension

import (
	"testing"

	"github.com/rprtr258/flatnotes/internal/goldmark"
	"github.com/rprtr258/flatnotes/internal/goldmark/renderer/html"
	"github.com/rprtr258/flatnotes/internal/goldmark/testutil"
)

func TestDefinitionList(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			DefinitionList,
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/definition_list.txt", t, testutil.ParseCliCaseArg()...)
}
