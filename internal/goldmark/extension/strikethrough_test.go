package extension

import (
	"testing"

	"github.com/rprtr258/flatnotes/internal/goldmark"
	"github.com/rprtr258/flatnotes/internal/goldmark/renderer/html"
	"github.com/rprtr258/flatnotes/internal/goldmark/testutil"
)

func TestStrikethrough(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			Strikethrough,
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/strikethrough.txt", t, testutil.ParseCliCaseArg()...)
}
