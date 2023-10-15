package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func ospathexists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

const MARKDOWN_EXT = ".md"

type InvalidTitleError struct {
	message string
}

func (e InvalidTitleError) Error() string {
	return fmt.Sprintf("specified title is invalid: %q", e.message)
}

type NoteDocument struct {
	Title   string
	Content string
	Tags    []string
	Modtime time.Time
}

func (self NoteDocument) ID() string {
	return self.Title
}

var _reImageBase64 = regexp.MustCompile(`!\[[^\[\]]*\]\(data:image/\w+;base64,[a-zA-Z0-9+/=]+\)`)

func (self NoteDocument) Fields() map[string]string {
	return map[string]string{
		"Title":   self.Title,
		"Content": _reImageBase64.ReplaceAllString(self.Content, ""),
		"Tags":    strings.Join(self.Tags, " "),
	}
}

func (self NoteDocument) note() Note {
	return Note{
		Title: self.Title,
		Tags:  self.Tags,
	}
}

type Note struct {
	Title    string
	NotesDir string
	Tags     []string // TODO: remove
}

func NewNote(notesDir, title string, new bool) (Note, error) {
	self := Note{
		Title:    strings.TrimSpace(title),
		NotesDir: notesDir,
		Tags:     nil,
	}
	if !_is_valid_title(self.Title) {
		return Note{}, fmt.Errorf("InvalidTitleError")
	}

	filepath := noteFilepath(notesDir, self.Title)
	if new && ospathexists(filepath) {
		return Note{}, fmt.Errorf("FileExistsError")
	}

	if new {
		f, err := os.Create(filepath)
		if err != nil {
			return Note{}, err
		}

		return self, f.Close()
	}

	return self, nil
}

func noteFilepath(dir, title string) string {
	return filepath.Join(dir, title+MARKDOWN_EXT)
}

func (self Note) Document() (NoteDocument, error) {
	content, err := self.GetContent()
	if err != nil {
		return NoteDocument{}, fmt.Errorf("get content %q: %w", self.Title, err)
	}

	modtime, err := self.LastModified()
	if err != nil {
		return NoteDocument{}, fmt.Errorf("get last modified time %q: %w", self.Title, err)
	}

	return NoteDocument{
		Title:   self.Title,
		Content: string(content),
		Tags:    self.Tags,
		Modtime: modtime,
	}, nil
}

func (self Note) LastModified() (time.Time, error) {
	filepath := noteFilepath(self.NotesDir, self.Title)
	stat, err := os.Stat(filepath)
	if err != nil {
		return time.Time{}, fmt.Errorf("get last modified time %q: %w", filepath, err)
	}

	return stat.ModTime(), nil
}

// Editable Properties
func (self *Note) SetTitle(new_title string) error {
	new_title = strings.TrimSpace(new_title)
	if !_is_valid_title(new_title) {
		return fmt.Errorf("InvalidTitleError")
	}

	oldTitle := self.Title
	self.Title = new_title
	err := os.Rename(
		noteFilepath(self.NotesDir, oldTitle),
		noteFilepath(self.NotesDir, new_title),
	)
	return err
}

func (self Note) GetContent() ([]byte, error) {
	return os.ReadFile(noteFilepath(self.NotesDir, self.Title))
}

func (self Note) SetContent(new_content []byte) error {
	filepath := noteFilepath(self.NotesDir, self.Title)
	if !ospathexists(filepath) {
		return fmt.Errorf("FileNotFoundError")
	}

	return os.WriteFile(filepath, new_content, 0o644)
}

func (self Note) Delete() error {
	return os.Remove(noteFilepath(self.NotesDir, self.Title))
}

// Return False if the declared title contains any of the following
// characters: <>:"/\|?*
func _is_valid_title(title string) bool {
	const invalid_chars = `<>:"/\|?*`
	return !strings.ContainsAny(title, invalid_chars)
}
