package extension

import (
	"testing"

	"github.com/rprtr258/flatnotes/internal/goldmark"
	"github.com/rprtr258/flatnotes/internal/goldmark/renderer/html"
	"github.com/rprtr258/flatnotes/internal/goldmark/testutil"
)

func TestTaskList(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			TaskList,
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/tasklist.txt", t, testutil.ParseCliCaseArg()...)
}