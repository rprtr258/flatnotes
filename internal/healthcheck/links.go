package healthcheck

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rprtr258/flatnotes/internal"
	"github.com/rprtr258/fun"
	"github.com/rprtr258/fun/iter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	gmext "github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// md is a reusable goldmark parser with the Linkify extension enabled,
// so bare URLs (https://example.com with no surrounding markup) are
// surfaced as autolink nodes alongside [text](url) and <url>.
var md = goldmark.New(goldmark.WithExtensions(gmext.Linkify))

// isAbsoluteLink reports whether url has a scheme or is protocol-relative.
func isAbsoluteLink(u string) bool {
	// _,err:=url.Parse(u)
	// return err==nil
	if strings.HasPrefix(u, "//") {
		return true
	}
	scheme, _, found := strings.Cut(u, ":")
	if !found || scheme == "" {
		return false
	}
	for _, r := range scheme {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') &&
			(r < '0' || r > '9') && r != '+' && r != '-' && r != '.' {
			return false
		}
	}
	return true
}

type noteLink struct {
	note string
	text string
	url  string
}

// collectLinks parses content as markdown via goldmark and appends every
// link found to out. Covered forms (all CommonMark-compliant, so
// balanced parens in URLs, escaped chars, etc. are handled by the parser):
//
//   - inline links [text](url)
//   - reference-style link definitions [ref]: url (and their usages)
//   - angle-bracket autolinks <https://...>
//   - bare URLs (via the Linkify extension)
//
// Images ![alt](url) are a distinct node kind and intentionally excluded.
func collectLinks(content, note string) iter.Seq[noteLink] {
	src := []byte(content)
	doc := md.Parser().Parse(text.NewReader(src))
	return func(yield func(noteLink) bool) {
		_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if !entering {
				return ast.WalkContinue, nil
			}
			switch n.Kind() {
			case ast.KindLink:
				if !yield(noteLink{
					note: note,
					text: inlineText(n, src),
					url:  string(n.(*ast.Link).Destination),
				}) {
					return ast.WalkStop, nil
				}
			case ast.KindAutoLink:
				if !yield(noteLink{
					note: note,
					text: string(n.(*ast.AutoLink).URL(src)),
					url:  string(n.(*ast.AutoLink).URL(src)),
				}) {
					return ast.WalkStop, nil
				}
			case ast.KindLinkReferenceDefinition:
				if !yield(noteLink{
					note: note,
					text: string(n.(*ast.LinkReferenceDefinition).Label),
					url:  string(n.(*ast.LinkReferenceDefinition).Destination),
				}) {
					return ast.WalkStop, nil
				}
			}
			return ast.WalkContinue, nil
		})
	}
}

// inlineText returns the concatenated text of a node's child *ast.Text
// nodes — used to recover a link's label without the deprecated
// Node.Text method.
func inlineText(n ast.Node, src []byte) string {
	var b strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			b.Write(t.Segment.Value(src))
		}
	}
	return b.String()
}

// checkNoteLinks loops over every note and collects all links,
// classifying each as relative or absolute. Relative links are printed
// as-is; absolute (remote) links are fetched and only the unavailable
// ones are reported.
func checkNoteLinks(notes map[string]internal.NoteDocument, dir string) iter.Seq[diagnostic] {
	var links []noteLink
	for title, doc := range notes {
		for link := range collectLinks(doc.Content, title) {
			links = append(links, link)
		}
	}

	relatives, absolutes := 0, 0
	var absoluteLinks, relativeLinks []noteLink
	for _, l := range links {
		isAbsolute := isAbsoluteLink(l.url)
		*fun.IF(isAbsolute, &absolutes, &relatives)++
		if isAbsolute {
			absoluteLinks = append(absoluteLinks, l)
		} else {
			relativeLinks = append(relativeLinks, l)
		}
	}
	fmt.Printf("=== note links (%d) ===\n", len(links))
	fmt.Printf("%d absolute links\n", absolutes)
	fmt.Printf("%d relative links\n", relatives)

	return iter.Concat(
		checkRelativeLinks(dir, relativeLinks),
		// checkRemoteLinks(absoluteLinks),
	)
}

// checkRelativeLinks resolves each relative link against the directory of
// the note it appears in (notes live under app.Dir as <title>.md) and
// reports only those whose target file is missing or a directory. Path
// traversal outside app.Dir is treated as invalid. Resolving mirrors the
// /attachments route's logic.
func checkRelativeLinks(dir string, links []noteLink) iter.Seq[diagnostic] {
	return func(yield func(diagnostic) bool) {
		missing := 0
		for _, l := range links {
			baseDir, _ := filepath.Abs(filepath.Join(dir, fun.IF(l.url[0] == '/', "/", filepath.Dir(l.note))))
			rel, _ := url.PathUnescape(l.url)
			abs := filepath.Join(baseDir, filepath.FromSlash(rel))
			resolved, err := filepath.Rel(baseDir, abs)
			if err != nil || resolved == ".." || strings.HasPrefix(resolved, ".."+string(filepath.Separator)) {
				missing++
				if !yield(diagnostic{
					note:       l.note,
					text:       l.url,
					group:      "check_relative_links",
					diagnostic: "outside data dir",
				}) {
					return
				}
				continue
			}
			info, err := os.Stat(abs)
			if err != nil {
				missing++
				if !yield(diagnostic{
					note:       l.note,
					text:       l.url,
					group:      "check_relative_links",
					diagnostic: "missing",
				}) {
					return
				}
				continue
			}
			if info.IsDir() {
				missing++
				if !yield(diagnostic{
					note:       l.note,
					text:       l.url,
					group:      "check_relative_links",
					diagnostic: "is directory",
				}) {
					return
				}
			}
		}
		fmt.Printf("%d missing / %d relative links\n", missing, len(links))
	}
}

// httpClient is a shared client with a short timeout for probing remote
// links. HEAD is sent first (cheap); servers that reject it fall back to
// GET. A link is unavailable if the request errors or returns a status
// >= 400.
var httpClient = &http.Client{Timeout: 10 * time.Second}

const (
	_tries      = 5
	_retryDelay = 5 * time.Second
	_threads    = 50
)

// Cache for probed URL status codes. One file per URL (sha256-hex of the URL);
// freshness is the file's mtime. A fresh cache hit skips the probe.
const (
	_cacheTTL = 24 * time.Hour
	_cacheDir = "/tmp/flatnotes-linkcheck"
)

// checkRemoteLinks probes each link concurrently with workers and
// prints the unavailable ones (status >= 400 or request error).
func checkRemoteLinks(links []noteLink) {
	fmt.Printf("=== remote link availability (%d) ===\n", len(links))
	codes := make([]int, len(links))
	var wg sync.WaitGroup
	jobs := make(chan int, _threads)
	wg.Add(_threads)
	var progress sync.Mutex
	for range _threads {
		go func() {
			defer wg.Done()
			for i := range jobs {
				codes[i] = isAvailable(links[i].url)
				progress.Lock()
				progress.Unlock()
			}
		}()
	}
	for i := range links {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	unavailable := 0
	for i, l := range links {
		if codes[i] < 400 {
			continue
		}
		unavailable++
		fmt.Printf("[unavailable] %s: %q -> %d\n", l.note, l.url, codes[i])
	}
	fmt.Printf("%d unavailable / %d remote links\n", unavailable, len(links))
}

func isAvailable(url string) int {
	code, ok := getOr(url, func() (int, error) {
		r, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return 0, err
		}
		r.Header.Add("dnt", "1")
		r.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36")
		r.Header.Add("sec-fetch-mode", "navigate")

		resp, err := httpClient.Do(r)
		if err != nil {
			return 0, err
		}
		for range _tries {
			if resp.StatusCode != http.StatusTooManyRequests {
				break
			}
			resp.Body.Close()
			time.Sleep(_retryDelay + rand.N(5*time.Second))
			resp, err = httpClient.Get(url)
			if err != nil {
				return 0, err
			}
		}
		defer resp.Body.Close()
		code := resp.StatusCode
		if code < 400 {
			return code, nil
		}
		return 0, fmt.Errorf("code %d", code)
	})
	return fun.IF(ok, code, 999)
}
