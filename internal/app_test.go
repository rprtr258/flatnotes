package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSearchHighlightShortLine exercises the search-highlight path where a
// match ("<mark>") appears within the first 100 runes of a content line.
// Previously postProcessHighlight computed substring(line, j-100, 300) with a
// negative offset, panicking: "runtime error: slice bounds out of range [-N:]".
// TestSearchHighlightMultibyteContent exercises the search-highlight path
// where content before the match contains multibyte UTF-8 runes.
func TestSearchHighlightMultibyteContent(t *testing.T) {
	dir := t.TempDir()
	// many multibyte runes before the match so the byte offset of "hello"
	// is well past the rune count of the context window.
	content := strings.Repeat("€", 200) + " hello world"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "note.md"), []byte(content), 0o644))

	app := use(New(dir))(t)

	use(app.Search("hello", SortScore, OrderNone, 0))
}

func TestSearchHighlightShortLine(t *testing.T) {
	dir := t.TempDir()
	// content line is shorter than 100 runes and the match starts at index 0,
	// so j-100 == -100 -> substring must clamp the offset instead of panicking.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "note.md"), []byte("hello world"), 0o644))

	app := use(New(dir))(t)

	use(app.Search("hello", SortScore, OrderNone, 0))
}
