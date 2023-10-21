package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rprtr258/flatnotes/internal/fts"
	"github.com/samber/lo"
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
	Tags    Set[string]
	Modtime time.Time
}

func (d NoteDocument) ID() string {
	return d.Title
}

var _reImageBase64 = regexp.MustCompile(`!\[[^\[\]]*\]\(data:image/\w+;base64,[a-zA-Z0-9+/=]+\)`)

func (d NoteDocument) Fields() map[string]fts.DocumentField {
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
			Content: strings.Join(lo.Keys(d.Tags), " "),
			Weight:  4,
			Terms:   lo.Keys(d.Tags),
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

func createNote(dir, title, content string) (Note, time.Time, error) {
	note := Note{
		Title:    title,
		NotesDir: dir,
	}

	filepath := noteFilepath(dir, note.Title)

	f, err := os.Create(filepath)
	if err != nil {
		if os.IsExist(err) {
			return Note{}, time.Time{}, fmt.Errorf("file exists: %q", filepath)
		}

		return Note{}, time.Time{}, err
	}
	defer f.Close()

	if _, err := f.Write([]byte(content)); err != nil {
		return Note{}, time.Time{}, fmt.Errorf("write content: %w", err)
	}

	stat, err := f.Stat()
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

	_, tags := extractTags(string(content))

	modtime, err := note.LastModified()
	if err != nil {
		return NoteDocument{}, fmt.Errorf("get last modified time %q: %w", note.Title, err)
	}

	return NoteDocument{
		Title:   note.Title,
		Content: string(content),
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
	n.Title = newTitle
	if err := os.Rename(
		noteFilepath(n.NotesDir, oldTitle),
		noteFilepath(n.NotesDir, newTitle),
	); err != nil {
		return fmt.Errorf("rename %q to %q: %w", oldTitle, newTitle, err)
	}

	return nil
}

func (n Note) GetContent() ([]byte, error) {
	return os.ReadFile(noteFilepath(n.NotesDir, n.Title))
}

func (n Note) SetContent(newContent []byte) error {
	filepath := noteFilepath(n.NotesDir, n.Title)
	if !ospathexists(filepath) {
		return fmt.Errorf("FileNotFoundError")
	}

	return os.WriteFile(filepath, newContent, 0o644)
}

func (n Note) Delete() error {
	return os.Remove(noteFilepath(n.NotesDir, n.Title))
}
