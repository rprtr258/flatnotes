package internal

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/rprtr258/flatnotes/internal/fts"
	"github.com/rprtr258/fun"
	"github.com/rprtr258/fun/set"
)

var (
	ErrTitleExists  = errors.New("the specified title already exists")
	ErrTitleInvalid = errors.New("the specified title contains invalid characters")
	ErrNotFound     = errors.New("the specified note cannot be found")
)

var (
	_reTags       = regexp.MustCompile(`(?:^#|\s#)(\w+)(?:\s|$)`)
	_reCodeblocks = regexp.MustCompile("`{1,3}.*?`{1,3}" /*, re.DOTALL*/)
)

// Return False if the declared title contains any of the following
// characters: <>:"/\|?*
func isValidTitle(title string) bool {
	const _invalidChars = `<>:"/\|?*` + "\n\r\t"
	return !strings.ContainsAny(title, _invalidChars)
}

// substring return part of a string
func substring(str string, offset, length int) string {
	return string(fun.Subslice(offset, length, []rune(str)...))
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
	tagsSet := set.New[string](0)
	for _, tag := range tags {
		tagsSet.Add(strings.ToLower(tag))
	}
	return contentExTags, tagsSet
}

func stripExt(filename string) string {
	_, fname := filepath.Split(filename)
	name, _ := strings.CutSuffix(fname, _markdownExt)
	return name
}

type App struct {
	Dir   string
	Index *fts.Index[NoteDocument]
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
	log.Info().Str("duration", time.Since(start).String()).Msg("finished initial indexing")

	return res, nil
}

type SearchResult struct {
	Note
	Score                              float64
	TitleHighlights, ContentHighlights string
	TagMatches                         []string
}

func (app *App) newSearchResult(hit fts.Hit[NoteDocument]) (SearchResult, error) {
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
			j := strings.Index(line, "<mark>")
			lines[i] = substring(line, j-100, 300)
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

// Return a list containing a Note object for every file in the notes
// directory.
func (app *App) getNotes() ([]Note, error) {
	matches, err := filepath.Glob(filepath.Join(app.Dir, "*"+_markdownExt))
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
// TODO: optimize
func (app *App) updateIndex() error {
	indexed := set.New[string](0)
	docs := []NoteDocument{}
	for id, doc := range app.Notes {
		idxFilename := id + _markdownExt
		idxFilepath := filepath.Join(app.Dir, idxFilename)
		if _, err := os.Stat(idxFilepath); os.IsNotExist(err) {
			// Delete missing
			app.Index.Delete(doc.ID())
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
func (app *App) GetTags() (set.Set[string], error) {
	if err := app.updateIndex(); err != nil {
		return set.Set[string]{}, err
	}

	res := set.New[string](0)
	for _, note := range app.Notes {
		res.Merge(note.Tags)
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

type SearchResultModel struct {
	Note              NoteDocument
	SearchResult      SearchResult
	LastModified      time.Time
	TitleHighlights   string
	ContentHighlights string
	TagMatches        []string
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

	var hits []fts.Hit[NoteDocument]
	// Parse Query
	if phrase == "*" {
		hits = fun.MapToSlice(app.Notes, func(_ string, doc NoteDocument) fts.Hit[NoteDocument] {
			return fts.Hit[NoteDocument]{
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

	slices.SortFunc(hits, func(i, j fts.Hit[NoteDocument]) int {
		if i.Score != j.Score {
			return cmp.Compare(j.Score, i.Score)
		}

		return cmp.Compare(app.Notes[j.ID].Modtime.Unix(), app.Notes[i.ID].Modtime.Unix())
	})

	if limit > 0 {
		hits = fun.Subslice(0, limit, hits...)
	}

	return fun.MapErr[SearchResultModel, fts.Hit[NoteDocument], error](
		func(hit fts.Hit[NoteDocument]) (SearchResultModel, error) {
			searchRes, err := app.newSearchResult(hit)
			if err != nil {
				return SearchResultModel{}, fmt.Errorf("map search result %v: %w", hit, err)
			}

			modtime, err := searchRes.LastModified()
			if err != nil {
				return SearchResultModel{}, fmt.Errorf("get last modified time %q: %w", searchRes.Title, err)
			}

			return SearchResultModel{
				Note:              app.Notes[hit.ID],
				SearchResult:      searchRes,
				LastModified:      modtime,
				TitleHighlights:   searchRes.TitleHighlights,
				ContentHighlights: searchRes.ContentHighlights,
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

func (app *App) CreateNote(data NotePostModel) (NoteContentResponseModel, error) {
	if !isValidTitle(data.Title) {
		return NoteContentResponseModel{}, ErrTitleInvalid
	}

	note, lastModified, err := createNote(app.Dir, data.Title, data.Content)
	if err != nil {
		return NoteContentResponseModel{}, err
	}

	return NoteContentResponseModel{
		NoteResponseModel: NoteResponseModel{
			Title:        note.Title,
			LastModified: lastModified.Unix(),
		},
		Content: data.Content,
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
