package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rprtr258/fun/set"

	"github.com/rprtr258/flatnotes/internal/fts"
)

func ospathexists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

const _markdownExt = ".md"

type InvalidTitleError struct {
	message string
}

func (e InvalidTitleError) Error() string {
	return fmt.Sprintf("specified title is invalid: %q", e.message)
}

type NoteDocument struct {
	Title   string
	Content string
	Tags    set.Set[string]
	Modtime time.Time
}

func (d NoteDocument) ID() string {
	return d.Title
}

var _reImageBase64 = regexp.MustCompile(`!\[[^\[\]]*\]\(data:image/\w+;base64,[a-zA-Z0-9+/=]+\)`)

func (d NoteDocument) Fields() map[string]fts.DocumentField {
	tags := d.Tags.List()
	return map[string]fts.DocumentField{
		"Title": {
			Content: d.Title,
			Weight:  2,
		},
		"Content": {
			Content: _reImageBase64.ReplaceAllString(d.Content, ""),
			Weight:  1,
		},
		"Tags": {
			Content: strings.Join(tags, " "),
			Weight:  4,
			Terms:   tags,
		},
	}
}

type Note struct {
	Title    string
	NotesDir string
}

func noteFilepath(dir, title string) string {
	return filepath.Join(dir, title+_markdownExt)
}

// removeEmptyParents removes empty directories from parent up to (but not
// including) dir, left behind after a note is moved or deleted.
func removeEmptyParents(dir, parent string) {
	for {
		rel, err := filepath.Rel(dir, parent)
		if err != nil || rel == "." || rel == "" || strings.HasPrefix(rel, "..") {
			return
		}
		if err := os.Remove(parent); err != nil {
			return // not empty or missing — stop
		}
		parent = filepath.Dir(parent)
	}
}

func createNote(dir, title, content string) (Note, time.Time, error) {
	note := Note{
		Title:    title,
		NotesDir: dir,
	}

	notePath := noteFilepath(dir, note.Title)

	if err := os.MkdirAll(filepath.Dir(notePath), 0o755); err != nil {
		return Note{}, time.Time{}, fmt.Errorf("create dirs: %w", err)
	}

	noteFile, err := os.OpenFile(notePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o666)
	if err != nil {
		if os.IsExist(err) {
			return Note{}, time.Time{}, ErrTitleExists
		}

		return Note{}, time.Time{}, err
	}
	defer noteFile.Close()

	if _, err := noteFile.WriteString(content); err != nil {
		return Note{}, time.Time{}, fmt.Errorf("write content: %w", err)
	}

	stat, err := noteFile.Stat()
	if err != nil {
		return Note{}, time.Time{}, fmt.Errorf("stat: %w", err)
	}

	lastModified := stat.ModTime()
	return note, lastModified, nil
}

func toDocument(note Note) (NoteDocument, error) {
	content, err := note.GetContent()
	if err != nil {
		return NoteDocument{}, fmt.Errorf("get content %q: %w", note.Title, err)
	}

	_, tags := extractTags(content)

	modtime, err := note.LastModified()
	if err != nil {
		return NoteDocument{}, fmt.Errorf("get last modified time %q: %w", note.Title, err)
	}

	return NoteDocument{
		Title:   note.Title,
		Content: content,
		Tags:    tags,
		Modtime: modtime,
	}, nil
}

func (n Note) LastModified() (time.Time, error) {
	filepath := noteFilepath(n.NotesDir, n.Title)
	stat, err := os.Stat(filepath)
	if err != nil {
		return time.Time{}, fmt.Errorf("get last modified time %q: %w", filepath, err)
	}

	return stat.ModTime(), nil
}

// Editable Properties
func (n *Note) SetTitle(newTitle string) error {
	oldTitle := n.Title
	oldFilepath := noteFilepath(n.NotesDir, oldTitle)
	newFilepath := noteFilepath(n.NotesDir, newTitle)

	if err := os.MkdirAll(filepath.Dir(newFilepath), 0o755); err != nil {
		return fmt.Errorf("create dirs: %w", err)
	}

	n.Title = newTitle
	if err := os.Rename(oldFilepath, newFilepath); err != nil {
		n.Title = oldTitle
		return fmt.Errorf("rename %q to %q: %w", oldTitle, newTitle, err)
	}

	removeEmptyParents(n.NotesDir, filepath.Dir(oldFilepath))
	return nil
}

func (n Note) GetContent() (string, error) {
	data, err := os.ReadFile(noteFilepath(n.NotesDir, n.Title))
	return string(data), err
}

func (n Note) SetContent(newContent []byte) error {
	filepath := noteFilepath(n.NotesDir, n.Title)
	if !ospathexists(filepath) {
		return fmt.Errorf("FileNotFoundError")
	}

	return os.WriteFile(filepath, newContent, 0o644)
}

func (n Note) Delete() error {
	notePath := noteFilepath(n.NotesDir, n.Title)
	if err := os.Remove(notePath); err != nil {
		return err
	}
	removeEmptyParents(n.NotesDir, filepath.Dir(notePath))
	return nil
}
