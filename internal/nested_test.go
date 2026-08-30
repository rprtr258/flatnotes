package internal

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNestedNotes(t *testing.T) {
	dir := t.TempDir()
	app, err := New(dir)
	require.NoError(t, err)

	// nested title is now valid
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

	// create nested note
	_, err = app.CreateNote("projects/flatnotes/ideas", "# ideas\n#tag")
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(dir, "projects", "flatnotes", "ideas.md"))

	// it appears in getNotes (recursive)
	notes, err := app.getNotes()
	require.NoError(t, err)
	require.Len(t, notes, 1)
	require.Equal(t, "projects/flatnotes/ideas", notes[0].Title)

	// read it back
	got, err := app.GetNote("projects/flatnotes/ideas")
	require.NoError(t, err)
	require.Equal(t, "projects/flatnotes/ideas", got.Title)
	require.Contains(t, got.Content, "ideas")

	// index sees it (lazily refreshed)
	require.NoError(t, app.updateIndex())
	require.Contains(t, app.Notes, "projects/flatnotes/ideas")

	// rename across a new nested path
	_, err = app.UpdateNote("projects/flatnotes/ideas", NotePatchModel{NewTitle: new("archive/old/ideas2")})
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(dir, "archive", "old", "ideas2.md"))
	// old empty dirs cleaned up
	require.NoDirExists(t, filepath.Join(dir, "projects"))

	// delete cleans empty parent dirs up to dir
	require.NoError(t, app.DeleteNote("archive/old/ideas2"))
	require.NoDirExists(t, filepath.Join(dir, "archive"))
}

func TestNestedSearch(t *testing.T) {
	dir := t.TempDir()
	app, err := New(dir)
	require.NoError(t, err)
	_, err = app.CreateNote("docs/readme", "searchable needle here")
	require.NoError(t, err)
	res, err := app.Search("needle", SortScore, OrderNone, 0)
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "docs/readme", res[0].Title)
}
