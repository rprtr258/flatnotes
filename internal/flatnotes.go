package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/rprtr258/flatnotes/internal/fts"
	"github.com/samber/lo"
)

type IndexSchema struct {
	// filename = ID(unique=True, stored=True)
	// last_modified = DATETIME(stored=True, sortable=True)
	// title = TEXT(
	//     field_boost=2.0, analyzer=StemmingFoldingAnalyzer, sortable=True
	// )
	// content = TEXT(analyzer=StemmingFoldingAnalyzer)
	// tags = KEYWORD(lowercase=True, field_boost=2.0)
}

type SearchResult struct {
	Note
	Score                                          float64
	TitleHighlights, ContentHighlights, TagMatches string
}

func NewSearchResult(notesDir string, hit fts.Hit[NoteDocument]) (SearchResult, error) {
	note, err := NewNote(notesDir, hit.Doc.ID(), false)
	if err != nil {
		return SearchResult{}, fmt.Errorf("get note %q: %w", hit.Doc.ID(), err)
	}

	// If the search was ordered using a text field then hit.score is the
	// value of that field. This isn't useful so only set self._score if it
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

var (
	TAGS_RE      = regexp.MustCompile(`(?:^#|\s#)(\w+)(?:\s|$)`)
	CODEBLOCK_RE = regexp.MustCompile("`{1,3}.*?`{1,3}" /*, re.DOTALL*/)
)

// Strip tags from the given content and return a tuple consisting of:
//
// - The content without the tags.
// - A set of tags converted to lowercase.
func extract_tags(content string) (string, Set[string]) {
	content_ex_codeblock := CODEBLOCK_RE.ReplaceAllLiteralString(content, "")
	_, tags := re_extract(TAGS_RE, content_ex_codeblock)
	content_ex_tags, _ := re_extract(TAGS_RE, content)
	// try {
	tagsSet := Set[string]{}
	for _, tag := range tags {
		tagsSet[strings.ToLower(tag)] = struct{}{}
	}
	return content_ex_tags, tagsSet
	// } except IndexError{
	// return content, set()
	// }
}

type Flatnotes struct {
	dir   string
	index *fts.Index[NoteDocument]
}

func NewFlatnotes(dir string) (Flatnotes, error) {
	if stat, err := os.Stat(dir); os.IsNotExist(err) {
		return Flatnotes{}, fmt.Errorf("not a directory: %q does not exist", dir)
	} else if !stat.IsDir() {
		return Flatnotes{}, fmt.Errorf("not a directory: %q is not a directory", dir)
	}

	self := Flatnotes{
		dir: dir,
	}

	// for now loaded from fs on startup
	self.index = fts.NewIndex[NoteDocument]()

	if err := self.update_index(); err != nil {
		return Flatnotes{}, fmt.Errorf("update index: %w", err)
	}

	return self, nil
}

// Load the note index or create new if not exists.
// Add a Note object to the index using the given writer. If the
// filename already exists in the index an update will be performed
// instead.
func (self *Flatnotes) _add_note_to_index(note Note) error {
	doc, err := note.Document()
	if err != nil {
		return fmt.Errorf("get document: %w", err)
	}

	self.index.Add(doc)
	return nil
}

// Return a list containing a Note object for every file in the notes
// directory.
func (self *Flatnotes) _get_notes() ([]Note, error) {
	matches, err := filepath.Glob(filepath.Join(self.dir, "*"+MARKDOWN_EXT))
	if err != nil {
		return nil, fmt.Errorf("glob: %w", err)
	}

	res := []Note{}
	for _, match := range matches {
		_, file := filepath.Split(match)
		note, err := NewNote(self.dir, strip_ext(file), false)
		if err != nil {
			return nil, fmt.Errorf("new note %q: %w", file, err)
		}

		res = append(res, note)
	}
	return res, nil
}

// Synchronize the index with the notes directory.
// Specify clean=True to completely rebuild the index
func (self *Flatnotes) update_index() error {
	indexed := Set[string]{}
	for id, doc := range self.index.Documents {
		idx_filename := id + MARKDOWN_EXT
		idx_filepath := filepath.Join(self.dir, idx_filename)
		if _, err := os.Stat(idx_filepath); os.IsNotExist(err) {
			// Delete missing
			self.index.Remove(id)
			log.Println(id, "removed from index")
		} else if stat, err := os.Stat(idx_filepath); err == nil && stat.ModTime().After(doc.Modtime) {
			note, err := NewNote(self.dir, id, false)
			if err != nil {
				return fmt.Errorf("get note %q: %w", id, err)
			}

			if err := self._add_note_to_index(note); err != nil {
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
	notes, err := self._get_notes()
	if err != nil {
		return fmt.Errorf("get notes: %w", err)
	}

	for _, note := range notes {
		if !indexed.Has(note.Title) {
			if err := self._add_note_to_index(note); err != nil {
				return fmt.Errorf("add note %q to index: %w", note.Title, err)
			}

			log.Printf("%q added to index\n", note.Title)
		}
	}
	return nil
}

// Return a list of all indexed tags.
func (self *Flatnotes) GetTags() ([]string, error) {
	return nil, self.update_index()
	// return self.index.field_terms("tags")
}

func (self *Flatnotes) pre_process_search_term(term string) string {
	term = strings.TrimSpace(term)
	// Replace "#tagname" with "tags:tagname"
	// term = TAGS_RE.ReplaceAllStringFunc(term, func(s string) string {
	// 	return "tags:" + s[1:]
	// })
	return term
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
func (self *Flatnotes) Search(
	phrase string,
	sortt Sort,
	order Order,
	limit int,
) ([]SearchResult, error) {
	if err := self.update_index(); err != nil {
		return nil, fmt.Errorf("update index: %w", err)
	}

	phrase = self.pre_process_search_term(phrase)

	var hits []fts.Hit[NoteDocument]
	// Parse Query
	if phrase == "*" {
		hits = lo.MapToSlice(self.index.Documents, func(_ string, doc NoteDocument) fts.Hit[NoteDocument] {
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
		hits = self.index.Search(
			phrase,
			// /*sortedby=*/ sort,
			// /*reverse=*/ reverse,
			// /*limit=*/ limit,
			// /*terms=*/ true,
		)
	}

	sort.Slice(hits, func(i, j int) bool {
		return hits[i].Doc.Modtime.After(hits[j].Doc.Modtime)
	})

	if limit > 0 {
		hits = lo.Slice(hits, 0, limit)
	}

	res := []SearchResult{}
	for _, hit := range hits {
		searchRes, err := NewSearchResult(self.dir, hit)
		if err != nil {
			return nil, fmt.Errorf("map search result %v: %w", hit, err)
		}

		res = append(res, searchRes)
	}
	return res, nil
}
