package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/samber/lo"

	"github.com/rprtr258/flatnotes/internal/fts"
)

var (
	ErrTitleExists  = fmt.Errorf("The specified title already exists.")
	ErrTitleInvalid = fmt.Errorf("The specified title contains invalid characters.")
	ErrNotFound     = fmt.Errorf("The specified note cannot be found.")
)

var (
	_reTags       = regexp.MustCompile(`(?:^#|\s#)(\w+)(?:\s|$)`)
	_reCodeblocks = regexp.MustCompile("`{1,3}.*?`{1,3}" /*, re.DOTALL*/)
)

// Return False if the declared title contains any of the following
// characters: <>:"/\|?*
func isValidTitle(title string) bool {
	const _invalidChars = ` <>:"/\|?*` + "\n\r\t"
	return !strings.ContainsAny(title, _invalidChars)
}

// Similar to re.sub but returns a tuple of:
//
// - `string` with matches removed
// - list of matches
func reExtract(re *regexp.Regexp, s string) (string, []string) {
	text := re.ReplaceAllLiteralString(s, "")
	matches := re.FindAllStringSubmatch(s, -1)
	return text, lo.Map(matches, func(match []string, _ int) string {
		return match[1]
	})
}

// Strip tags from the given content and return a tuple consisting of:
//
// - The content without the tags.
// - A set of tags converted to lowercase.
func extractTags(content string) (string, Set[string]) {
	contentExCodeblock := _reCodeblocks.ReplaceAllLiteralString(content, "")
	_, tags := reExtract(_reTags, contentExCodeblock)
	contentExTags, _ := reExtract(_reTags, content)
	tagsSet := Set[string]{}
	for _, tag := range tags {
		tagsSet[strings.ToLower(tag)] = struct{}{}
	}
	return contentExTags, tagsSet
}

func stripExt(filename string) string {
	_, fname := filepath.Split(filename)
	name, _ := strings.CutSuffix(fname, _markdownExt)
	return name
}

type App struct {
	dir   string
	index *fts.Index[NoteDocument]
}

func New(dir string) (App, error) {
	if stat, err := os.Stat(dir); os.IsNotExist(err) {
		return App{}, fmt.Errorf("not a directory: %q does not exist", dir)
	} else if !stat.IsDir() {
		return App{}, fmt.Errorf("not a directory: %q is not a directory", dir)
	}

	res := App{
		dir:   dir,
		index: fts.NewIndex[NoteDocument](),
	}

	// for now loaded from fs on startup
	if err := res.updateIndex(); err != nil {
		return App{}, fmt.Errorf("update index: %w", err)
	}

	return res, nil
}

type SearchResult struct {
	Note
	Score                                          float64
	TitleHighlights, ContentHighlights, TagMatches string
}

func (app *App) newSearchResult(hit fts.Hit[NoteDocument]) (SearchResult, error) {
	note, err := app.getNote(hit.Doc.ID())
	if err != nil {
		return SearchResult{}, fmt.Errorf("get note %q: %w", hit.Doc.ID(), err)
	}

	// If the search was ordered using a text field then hit.score is the
	// value of that field. This isn't useful so only set _score if it
	// is a float.

	var titleHighlights, contentHighlights, tagMatches string
	for _, field := range hit.Terms {
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(field.Term))
		// switch k {
		// case "Title":
		// 	titleHighlights += strings.Join(field, "\n")
		// case "Content":
		//	contentHighlights += strings.Join(field, "\n")
		contentHighlights += re.ReplaceAllStringFunc(hit.Doc.Content, func(s string) string {
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
		lines = lo.Filter(lines, func(line string, _ int) bool {
			return strings.Contains(line, "<mark>")
		})
		lines = lo.Slice(lines, 0, 3)
		return replacer.Replace(strings.Join(lines, "<br>"))
	}

	return SearchResult{
		Note:              note,
		Score:             hit.Score,
		TitleHighlights:   postProcessHighlight(titleHighlights),
		ContentHighlights: postProcessHighlight(contentHighlights),
		TagMatches:        postProcessHighlight(tagMatches),
	}, nil
}

// Load the note index or create new if not exists.
// Add a Note object to the index using the given writer. If the
// filename already exists in the index an update will be performed
// instead.
func (app *App) addNoteToIndex(note Note) error {
	doc, err := toDocument(note)
	if err != nil {
		return fmt.Errorf("get document: %w", err)
	}

	app.index.Add(doc)
	return nil
}

func (app *App) getNote(title string) (Note, error) {
	filepath := noteFilepath(app.dir, title)
	if !ospathexists(filepath) {
		return Note{}, ErrNotFound
	}

	// if !isValidTitle(title) {
	// 	return Note{}, ErrTitleInvalid
	// }

	return Note{
		// Title:    strings.TrimSpace(title),
		Title:    title,
		NotesDir: app.dir,
	}, nil
}

// Return a list containing a Note object for every file in the notes
// directory.
func (app *App) getNotes() ([]Note, error) {
	matches, err := filepath.Glob(filepath.Join(app.dir, "*"+_markdownExt))
	if err != nil {
		return nil, fmt.Errorf("glob: %w", err)
	}

	res := []Note{}
	for _, match := range matches {
		_, file := filepath.Split(match)
		note, err := app.getNote(stripExt(file))
		if err != nil {
			return nil, fmt.Errorf("new note %q: %w", file, err)
		}

		res = append(res, note)
	}
	return res, nil
}

// Synchronize the index with the notes directory.
// Specify clean=True to completely rebuild the index
func (app *App) updateIndex() error {
	indexed := Set[string]{}
	for id, doc := range app.index.Documents {
		idxFilename := id + _markdownExt
		idxFilepath := filepath.Join(app.dir, idxFilename)
		if _, err := os.Stat(idxFilepath); os.IsNotExist(err) {
			// Delete missing
			app.index.Remove(id)
			log.Println(id, "removed from index")
		} else if stat, err := os.Stat(idxFilepath); err == nil && stat.ModTime().After(doc.Modtime) {
			note, err := app.getNote(id)
			if err != nil {
				return fmt.Errorf("get note %q: %w", id, err)
			}

			if err := app.addNoteToIndex(note); err != nil {
				return fmt.Errorf("add note %q to index: %w", id, err)
			}

			// Update modified
			log.Println(id, "updated")

			indexed[id] = struct{}{}
		} else {
			// Ignore already indexed
			indexed[id] = struct{}{}
		}
	}

	// Add new
	notes, err := app.getNotes()
	if err != nil {
		return fmt.Errorf("get notes: %w", err)
	}

	for _, note := range notes {
		if !indexed.Has(note.Title) {
			if err := app.addNoteToIndex(note); err != nil {
				return fmt.Errorf("add note %q to index: %w", note.Title, err)
			}

			log.Printf("%q added to index\n", note.Title)
		}
	}
	return nil
}

// Return a list of all indexed tags.
func (app *App) GetTags() (Set[string], error) {
	if err := app.updateIndex(); err != nil {
		return nil, err
	}

	res := Set[string]{}
	for _, note := range app.index.Documents {
		for tag := range note.Tags {
			res[tag] = struct{}{}
		}
	}
	return res, nil
}

type Sort string

const (
	SortNone         Sort = ""
	SortScore        Sort = "score"
	SortTitle        Sort = "title"
	SortLastModified Sort = "last_modified"
)

type Order string

const (
	OrderNone Order = ""
	OrderAsc  Order = "asc"
	OrderDesc Order = "desc"
)

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

	var hits []fts.Hit[NoteDocument]
	// Parse Query
	if phrase == "*" {
		hits = lo.MapToSlice(app.index.Documents, func(_ string, doc NoteDocument) fts.Hit[NoteDocument] {
			return fts.Hit[NoteDocument]{
				Doc:   doc,
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
		hits = app.index.Search(
			phrase,
			// /*sortedby=*/ sort,
			// /*reverse=*/ reverse,
			// /*limit=*/ limit,
			// /*terms=*/ true,
			func() []string {
				_, tags := extractTags(phrase)
				return lo.Keys(tags)
			}(),
		)
	}

	sort.Slice(hits, func(i, j int) bool {
		return hits[i].Doc.Modtime.After(hits[j].Doc.Modtime)
	})

	if limit > 0 {
		hits = lo.Slice(hits, 0, limit)
	}

	res := []SearchResultModel{}
	for _, hit := range hits {
		searchRes, err := app.newSearchResult(hit)
		if err != nil {
			return nil, fmt.Errorf("map search result %v: %w", hit, err)
		}

		modtime, err := searchRes.LastModified()
		if err != nil {
			return nil, fmt.Errorf("get last modified time %q: %w", searchRes.Title, err)
		}

		toOption := func(s string) *string {
			if s == "" {
				return nil
			}
			return &s
		}
		res = append(res, SearchResultModel{
			Score:             searchRes.Score,
			Title:             searchRes.Title,
			LastModified:      modtime.Unix(),
			TitleHighlights:   toOption(searchRes.TitleHighlights),
			ContentHighlights: toOption(searchRes.ContentHighlights),
			TagMatches:        toOption(searchRes.TagMatches),
		})
	}
	return res, nil
}

func (app *App) GetNote(title string, includeContent bool) (NoteContentResponseModel, error) {
	note, err := app.getNote(title)
	if err != nil {
		return NoteContentResponseModel{}, fmt.Errorf("get note %q: %w", title, err)
	}

	modtime, err := note.LastModified()
	if err != nil {
		return NoteContentResponseModel{}, fmt.Errorf("get last modified time %q: %w", title, err)
	}

	resContent := (*string)(nil)
	if includeContent {
		content, err := note.GetContent()
		if err != nil {
			return NoteContentResponseModel{}, fmt.Errorf("get content: %w", err)
		}

		resContent = lo.ToPtr(string(content))
	}

	return NoteContentResponseModel{
		NoteResponseModel: NoteResponseModel{
			Title:        note.Title,
			LastModified: modtime.Unix(),
		},
		Content: resContent,
	}, nil
}

func (app *App) CreateNote(data NotePostModel) (NoteContentResponseModel, error) {
	if !isValidTitle(data.Title) {
		return NoteContentResponseModel{}, ErrTitleInvalid
	}

	note, lastModified, err := createNote(app.dir, data.Title, data.Content)
	if err != nil {
		return NoteContentResponseModel{}, err
	}

	return NoteContentResponseModel{
		NoteResponseModel: NoteResponseModel{
			Title:        note.Title,
			LastModified: lastModified.Unix(),
		},
		Content: &data.Content,
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
		Content: lo.ToPtr(doc.Content),
	}, nil
}

func (app *App) DeleteNote(title string) error {
	note, err := app.getNote(title)
	if err != nil {
		return err
	}

	return note.Delete()
}
