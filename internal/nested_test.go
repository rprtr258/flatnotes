package internal

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidTitle(t *testing.T) {
	for _, title := range []string{"folder/sub note"} {
		require.True(t, isValidTitle(title))
	}
	for _, title := range []string{
		"../etc",
		"a//b",
		"/leading",
		`back\slash`,
	} {
		require.False(t, isValidTitle(title))
	}
}

func TestNestedNotes(t *testing.T) {
	dir := t.TempDir()
	app := use(New(dir))(t)

	// create nested note
	use(app.CreateNote("projects/flatnotes/ideas", "# ideas\n#tag"))(t)
	require.FileExists(t, filepath.Join(dir, "projects", "flatnotes", "ideas.md"))

	// it appears in getNotes (recursive)
	notes := use(app.getNotes())(t)
	require.Len(t, notes, 1)
	require.Equal(t, "projects/flatnotes/ideas", notes[0].Title)

	// read it back
	got := use(app.GetNote("projects/flatnotes/ideas"))(t)
	require.Equal(t, "projects/flatnotes/ideas", got.Title)
	require.Contains(t, got.Content, "ideas")

	// index sees it (lazily refreshed)
	require.NoError(t, app.updateIndex())
	require.Contains(t, app.Notes, "projects/flatnotes/ideas")

	// rename across a new nested path
	use(app.UpdateNote("projects/flatnotes/ideas", NotePatchModel{NewTitle: new("archive/old/ideas2")}))(t)
	require.FileExists(t, filepath.Join(dir, "archive", "old", "ideas2.md"))
	// old empty dirs cleaned up
	require.NoDirExists(t, filepath.Join(dir, "projects"))

	// delete cleans empty parent dirs up to dir
	require.NoError(t, app.DeleteNote("archive/old/ideas2"))
	require.NoDirExists(t, filepath.Join(dir, "archive"))
}

func TestNestedSearch(t *testing.T) {
	dir := t.TempDir()

	app := use(New(dir))(t)

	use(app.CreateNote("docs/readme", "searchable needle here"))(t)
	res := use(app.Search("needle", SortScore, OrderNone, 0))(t)
	require.Len(t, res, 1)
	require.Equal(t, "docs/readme", res[0].Title)
}
