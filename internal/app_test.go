package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSearchHighlightShortLine exercises the search-highlight path where a
// match ("<mark>") appears within the first 100 runes of a content line.
// Previously postProcessHighlight computed substring(line, j-100, 300) with a
// negative offset, panicking: "runtime error: slice bounds out of range [-N:]".
func TestSearchHighlightShortLine(t *testing.T) {
	dir := t.TempDir()
	// content line is shorter than 100 runes and the match starts at index 0,
	// so j-100 == -100 -> substring must clamp the offset instead of panicking.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "note.md"), []byte("hello world"), 0o644))

	app, err := New(dir)
	require.NoError(t, err)

	require.NotPanics(t, func() {
		_, err = app.Search("hello", SortScore, OrderNone, 0)
	})
	require.NoError(t, err)
}
