package healthcheck

import (
	"fmt"
	"iter"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	gmext "github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"

	"github.com/rprtr258/flatnotes/internal"
	"github.com/rprtr258/fun"
)

// Run executes all health checks and prints results to stdout.
// Intended to be extended with more checks over time; each check is a
// function that prints its findings to the terminal.
func Run(app internal.App) {
	checkNoteLinks(app)
}

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
// classifying each as relative or absolute.
func checkNoteLinks(app internal.App) {
	var links []noteLink
	for title, doc := range app.Notes {
		for link := range collectLinks(doc.Content, title) {
			links = append(links, link)
		}
	}

	relatives, absolutes := 0, 0
	fmt.Printf("=== note links (%d) ===\n", len(links))
	for _, l := range links {
		isAbsolute := isAbsoluteLink(l.url)
		kind := fun.IF(isAbsolute, "absolute", "relative")
		*fun.IF(isAbsolute, &absolutes, &relatives)++
		if !isAbsolute {
			fmt.Printf("- [%s] %s: %q -> %s\n", kind, l.note, l.text, l.url)
		}
	}
	fmt.Printf("%d absolute links\n", absolutes)
	fmt.Printf("%d relative links\n", relatives)
}
