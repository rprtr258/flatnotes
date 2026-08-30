package internal

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
	"time"

	"github.com/rprtr258/fun"
	"github.com/rprtr258/fun/set"
	"github.com/rs/zerolog/log"

	"github.com/rprtr258/flatnotes/internal/fts"
)

var (
	ErrTitleExists  = errors.New("the specified title already exists")
	ErrTitleInvalid = errors.New("the specified title contains invalid characters")
	ErrNotFound     = errors.New("the specified note cannot be found")
)

var (
	_reTags       = regexp.MustCompile(`(?:^#|\s#)(\w+)(?:\s|$)`) // TODO: get from metadata
	_reCodeblocks = regexp.MustCompile("`{1,3}.*?`{1,3}" /*, re.DOTALL*/)
)

// isValidTitle reports whether title is a valid (possibly nested) note
// title. The '/' separator is allowed for nesting. Backslash, control chars,
// path-traversal ("..") and empty path segments are rejected.
func isValidTitle(title string) bool {
	if title == "" {
		return false
	}
	const _invalidChars = `<>:"\|?*` + "\n\r\t"
	if strings.ContainsAny(title, _invalidChars) {
		return false
	}
	for _, part := range strings.Split(title, "/") {
		if part == "" || part == ".." || part == "." {
			return false
		}
	}
	return true
}

// substring return part of a string
func substring(str string, offset, length int) string {
	r := []rune(str)
	offset = max(offset, 0)
	return string(r[offset:min(offset+length, len(r))])
}

// Similar to re.sub but returns a tuple of:
//
// - `string` with matches removed
// - list of matches
func reExtract(re *regexp.Regexp, s string) (string, []string) {
	text := re.ReplaceAllLiteralString(s, "")
	matches := re.FindAllStringSubmatch(s, -1)
	return text, fun.Map[string](func(match []string) string {
		return match[1]
	}, matches...)
}

// Strip tags from the given content and return a tuple consisting of:
//
// - The content without the tags.
// - A set of tags converted to lowercase.
func extractTags(content string) (string, set.Set[string]) {
	contentExCodeblock := _reCodeblocks.ReplaceAllLiteralString(content, "")
	_, tags := reExtract(_reTags, contentExCodeblock)
	contentExTags, _ := reExtract(_reTags, content)
	tagsSet := set.New[string](len(tags))
	for _, tag := range tags {
		tagsSet.Add(strings.ToLower(tag))
	}
	return contentExTags, tagsSet
}

type App struct {
	Dir   string
	Index *fts.Index[NoteDocument]
	// all Notes
	Notes map[string]NoteDocument
}

func New(dir string) (App, error) {
	if stat, err := os.Stat(dir); os.IsNotExist(err) {
		return App{}, fmt.Errorf("not a directory: %q does not exist", dir)
	} else if !stat.IsDir() {
		return App{}, fmt.Errorf("not a directory: %q is not a directory", dir)
	}

	res := App{
		Dir:   dir,
		Index: fts.NewIndex[NoteDocument](),
		Notes: map[string]NoteDocument{},
	}

	// for now loaded from fs on startup
	start := time.Now()
	log.Info().Msg("started initial indexing")
	if err := res.updateIndex(); err != nil {
		return App{}, fmt.Errorf("update index: %w", err)
	}
	log.Info().Dur("duration", time.Since(start)).Msg("finished initial indexing")

	return res, nil
}

type SearchResult struct {
	Note
	Score                              float64
	TitleHighlights, ContentHighlights string
	TagMatches                         []string
}

func (app *App) newSearchResult(hit fts.Hit) (SearchResult, error) {
	note, err := app.getNote(hit.ID)
	if err != nil {
		return SearchResult{}, fmt.Errorf("get note %q: %w", hit.ID, err)
	}

	// If the search was ordered using a text field then hit.score is the
	// value of that field. This isn't useful so only set _score if it
	// is a float.

	var titleHighlights, contentHighlights string
	for _, field := range hit.Terms {
		re := regexp.MustCompile(`\b(?i)` + regexp.QuoteMeta(field.Term) + `\b`)
		// switch k {
		// case "Title":
		// 	titleHighlights += strings.Join(field, "\n")
		// case "Content":
		//	contentHighlights += strings.Join(field, "\n")
		contentHighlights += re.ReplaceAllStringFunc(app.Notes[hit.ID].Content, func(s string) string {
			return "<mark>" + s + "</mark>"
		})
		// case "Tags":
		// 	tagMatches += strings.Join(field, "\n")
		// default:
		// 	log.Printf("unknown field: %v\n", field)
		// }
	}

	replacer := strings.NewReplacer(
		"<mark>", `<b class="match term0">`,
		"</mark>", `</b>`,
	)
	postProcessHighlight := func(s string) string {
		lines := strings.Split(s, "\n")
		lines = fun.Filter(func(line string) bool {
			return strings.Contains(line, "<mark>")
		}, lines...)
		lines = fun.Subslice(0, 3, lines...)
		for i, line := range lines {
			jBytes := strings.Index(line, "<mark>")
			jRunes := utf8.RuneCountInString(line[:jBytes])
			lines[i] = substring(line, jRunes-100, 300)
		}
		return replacer.Replace(strings.Join(lines, "<br>"))
	}

	return SearchResult{
		Note:              note,
		Score:             hit.Score,
		TitleHighlights:   postProcessHighlight(titleHighlights),
		ContentHighlights: postProcessHighlight(contentHighlights),
		TagMatches:        hit.Tags,
	}, nil
}

func (app *App) getNote(title string) (Note, error) {
	if !ospathexists(noteFilepath(app.Dir, title)) {
		return Note{}, ErrNotFound
	}

	return Note{
		Title:    title,
		NotesDir: app.Dir,
	}, nil
}

// Return a list containing a Note object for every markdown file in the
// notes directory, including those in nested subdirectories.
func (app *App) getNotes() ([]Note, error) {
	res := []Note{}
	err := filepath.WalkDir(app.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, _markdownExt) {
			return nil
		}
		rel, err := filepath.Rel(app.Dir, path)
		if err != nil {
			return fmt.Errorf("rel path %q: %w", path, err)
		}
		title := filepath.ToSlash(strings.TrimSuffix(rel, _markdownExt))
		note, err := app.getNote(title)
		if err != nil {
			return fmt.Errorf("new note %q: %w", title, err)
		}
		res = append(res, note)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk: %w", err)
	}
	return res, nil
}

// Synchronize the index with the notes directory.
// TODO: optimize
func (app *App) updateIndex() error {
	indexed := set.New[string](0)
	docs := []NoteDocument{}
	for id, doc := range app.Notes {
		idxFilename := id + _markdownExt
		idxFilepath := filepath.Join(app.Dir, idxFilename)
		if _, err := os.Stat(idxFilepath); os.IsNotExist(err) {
			// Delete missing
			app.Index.Delete(id)
			delete(app.Notes, id)
			log.Info().Str("id", id).Msg("removed from index")
		} else if stat, err := os.Stat(idxFilepath); err == nil && stat.ModTime().After(doc.Modtime) {
			note, err := app.getNote(id)
			if err != nil {
				return fmt.Errorf("get note %q: %w", id, err)
			}

			doc, err := toDocument(note)
			if err != nil {
				return fmt.Errorf("get document, %q: %w", note.Title, err)
			}

			docs = append(docs, doc)

			// Update modified
			log.Info().Str("id", id).Msg("updated")

			indexed.Add(id)
		} else {
			// Ignore already indexed
			indexed.Add(id)
		}
	}

	// Add new
	notes, err := app.getNotes()
	if err != nil {
		return fmt.Errorf("get notes: %w", err)
	}

	for _, note := range notes {
		if indexed.Contains(note.Title) {
			continue
		}

		doc, err := toDocument(note)
		if err != nil {
			return fmt.Errorf("get document, %q: %w", note.Title, err)
		}

		docs = append(docs, doc)
		log.Info().Str("title", note.Title).Msg("added to index")
	}

	app.Index.Add(docs...)
	for _, doc := range docs {
		app.Notes[doc.ID()] = doc
	}

	return nil
}

// Return a list of all indexed tags.
func (app *App) GetTags() ([]string, error) {
	if err := app.updateIndex(); err != nil {
		return nil, err
	}

	res := map[string]int{}
	for _, note := range app.Notes {
		for tag := range note.Tags.Iter() {
			res[tag]++
		}
	}

	entries := fun.Entries(res)
	slices.SortFunc(entries, func(a, b fun.Pair[string, int]) int {
		return b.V - a.V
	})
	return fun.Map[string](func(e fun.Pair[string, int]) string { return e.K }, entries...), nil
}

// taskListItemRE matches GFM task list items, e.g. "- [ ] todo" / "- [x] done".
var taskListItemRE = regexp.MustCompile(`(?m)^\s*[-*+]\s+\[([ xX])]\s+(.+)$`)

// GetTodos returns every task list item across all notes, grouped by note.
func (app *App) GetTodos() ([]NoteTodosModel, error) {
	if err := app.updateIndex(); err != nil {
		return nil, fmt.Errorf("update index: %w", err)
	}

	res := []NoteTodosModel{}
	for _, note := range app.Notes {
		todos := []TodoItemModel{}
		for _, m := range taskListItemRE.FindAllStringSubmatch(note.Content, -1) {
			todos = append(todos, TodoItemModel{
				Text: strings.TrimSpace(m[2]),
				Done: m[1] != " ",
			})
		}
		if len(todos) == 0 {
			continue
		}

		res = append(res, NoteTodosModel{
			Title:        note.Title,
			LastModified: note.Modtime.Unix(),
			Todos:        todos,
		})
	}

	slices.SortFunc(res, func(i, j NoteTodosModel) int {
		return cmp.Compare(i.Title, j.Title)
	})

	return res, nil
}

type Sort int

const (
	SortScore Sort = iota
	SortTitle
	SortLastModified
)

func (o Sort) String() string {
	return [...]string{
		"Score",
		"Title",
		"Last Modified",
	}[o]
}

type Order string

const (
	OrderNone Order = ""
	OrderAsc  Order = "asc"
	OrderDesc Order = "desc"
)

type SearchResultModel struct {
	Score             float64  `json:"score"`
	Title             string   `json:"title"`
	LastModified      int64    `json:"lastModified"`
	TitleHighlights   *string  `json:"titleHighlights"`
	ContentHighlights *string  `json:"contentHighlights"`
	TagMatches        []string `json:"tagMatches"`
}

// Search the index for the given term.
func (app *App) Search(
	phrase string,
	sortt Sort,
	order Order,
	limit int,
) ([]SearchResultModel, error) {
	if err := app.updateIndex(); err != nil {
		return nil, fmt.Errorf("update index: %w", err)
	}

	phrase = strings.TrimSpace(phrase)

	var hits []fts.Hit
	// Parse Query
	if phrase == "*" {
		hits = fun.MapToSlice(app.Notes, func(_ string, doc NoteDocument) fts.Hit {
			return fts.Hit{
				ID:    doc.ID(),
				Tags:  nil,
				Score: 0,
				Terms: nil,
			}
		})
	} else {
		// Determine Sort Direction
		// Note: Confusingly, when sorting by 'score', reverse = True means
		// asc so we have to flip the logic for that case!
		// reverse := order == "desc"
		// if sort == SortNone {
		// 	reverse = !reverse
		// }

		// Run Search
		hits = app.Index.Search(
			phrase,
			// /*sortedby=*/ sort,
			// /*reverse=*/ reverse,
			// /*limit=*/ limit,
			// /*terms=*/ true,
			func() []string {
				_, tags := extractTags(phrase)
				return tags.List()
			}(),
		)
	}

	slices.SortFunc(hits, func(i, j fts.Hit) int {
		if i.Score != j.Score {
			return cmp.Compare(j.Score, i.Score)
		}

		return cmp.Compare(app.Notes[j.ID].Modtime.Unix(), app.Notes[i.ID].Modtime.Unix())
	})

	if limit > 0 {
		hits = fun.Subslice(0, limit, hits...)
	}

	return fun.MapErr[SearchResultModel](func(hit fts.Hit) (SearchResultModel, error) {
		searchRes, err := app.newSearchResult(hit)
		if err != nil {
			return SearchResultModel{}, fmt.Errorf("map search result %v: %w", hit, err)
		}

		modtime, err := searchRes.LastModified()
		if err != nil {
			return SearchResultModel{}, fmt.Errorf("get last modified time %q: %w", searchRes.Title, err)
		}

		toOption := func(s string) *string {
			return fun.IF(s == "", nil, &s)
		}
		return SearchResultModel{
			Score:             searchRes.Score,
			Title:             searchRes.Title,
			LastModified:      modtime.Unix(),
			TitleHighlights:   toOption(searchRes.TitleHighlights),
			ContentHighlights: toOption(searchRes.ContentHighlights),
			TagMatches:        searchRes.TagMatches,
		}, nil
	}, hits...)
}

func (app *App) GetNote(title string) (NoteContentResponseModel, error) {
	note, err := app.getNote(title)
	if err != nil {
		return NoteContentResponseModel{}, err
	}

	modtime, err := note.LastModified()
	if err != nil {
		return NoteContentResponseModel{}, fmt.Errorf("get last modified time %q: %w", title, err)
	}

	content, err := note.GetContent()
	if err != nil {
		return NoteContentResponseModel{}, fmt.Errorf("get content: %w", err)
	}

	return NoteContentResponseModel{
		NoteResponseModel: NoteResponseModel{
			Title:        note.Title,
			LastModified: modtime.Unix(),
		},
		Content: content,
	}, nil
}

func (app *App) CreateNote(title, content string) (NoteContentResponseModel, error) {
	if !isValidTitle(title) {
		return NoteContentResponseModel{}, ErrTitleInvalid
	}

	note, lastModified, err := createNote(app.Dir, title, content)
	if err != nil {
		return NoteContentResponseModel{}, err
	}

	return NoteContentResponseModel{
		NoteResponseModel: NoteResponseModel{
			Title:        note.Title,
			LastModified: lastModified.Unix(),
		},
		Content: content,
	}, nil
}

func (app *App) UpdateNote(title string, data NotePatchModel) (NoteContentResponseModel, error) {
	if !isValidTitle(*data.NewTitle) {
		return NoteContentResponseModel{}, ErrTitleInvalid
	}

	note, err := app.getNote(title)
	if err != nil {
		return NoteContentResponseModel{}, fmt.Errorf("get note %q: %w", title, err)
	}

	if data.NewTitle != nil {
		if err := note.SetTitle(*data.NewTitle); err != nil {
			return NoteContentResponseModel{}, fmt.Errorf("set note %q title to %q: %w", title, *data.NewTitle, err)
		}
	}
	if data.NewContent != nil {
		if err := note.SetContent([]byte(*data.NewContent)); err != nil {
			return NoteContentResponseModel{}, fmt.Errorf("set note %q content: %w", title, err)
		}
	}

	doc, err := toDocument(note)
	if err != nil {
		return NoteContentResponseModel{}, fmt.Errorf("get note data %q: %w", title, err)
	}

	return NoteContentResponseModel{
		NoteResponseModel: NoteResponseModel{
			Title:        note.Title,
			LastModified: doc.Modtime.Unix(),
		},
		Content: doc.Content,
	}, nil
}

func (app *App) DeleteNote(title string) error {
	note, err := app.getNote(title)
	if err != nil {
		return err
	}

	return note.Delete()
}
