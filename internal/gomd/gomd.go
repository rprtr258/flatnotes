package gomd

import (
	"io"
	"strings"
)

// const Header = `<!DOCTYPE html>`

// WriteHTML writes escaped html to w
func WriteHTML(w io.Writer, s string) {
	b := []byte(s)
	last := 0
	for i, c := range b {
		seq, ok := map[byte]string{
			'\000': "\uFFFD",
			'"':    "&#34;", // shorter than "&quot;"
			'\'':   "&#39;", // shorter than "&apos;"
			'&':    "&amp;",
			'<':    "&lt;",
			'>':    "&gt;",
		}[c]
		if !ok {
			continue
		}
		w.Write(b[last:i])
		io.WriteString(w, seq)
		last = i + 1
	}
	w.Write(b[last:])
}

func DisplayMarkdown(w io.Writer, md string) {
	type Token struct {
		Type         int
		Start, End   string
		RStart, REnd string
	}

	const (
		None = iota
		Code
		H3
		H2
		H1
		List
		Mono
		Bold
		Italics
		Break
	)

	tokens := [...]Token{
		{Code, "```\r\n", "\r\n```\r\n", "<pre><code>", "</code></pre>"},
		{H3, "###", "\r\n", "<h6>", "</h6>"},
		{H2, "##", "\r\n", "<h5>", "</h5>"},
		{H1, "#", "\r\n", "<h4>", "</h4>"},
		{List, "\n-", "\r", `<li class="ms-4">`, "</li>"},
		{Mono, "`", "`", "<tt>", "</tt>"},
		{Bold, "**", "**", "<b>", "</b>"},
		{Bold, "__", "__", "<b>", "</b>"},
		{Italics, "*", "*", "<i>", "</i>"},
		{Italics, "_", "_", "<i>", "</i>"},
		{Break, "\r", "\n", "<br>", ""},
	}

	for len(md) > 0 {
		replaced := false
		for _, tok := range tokens {
			start := strings.Index(md, tok.Start)
			if start == -1 {
				continue
			}

			end := strings.Index(md[start+len(tok.Start):], tok.End)
			if end == -1 {
				continue
			}
			if end+len(tok.Start) == 0 {
				continue
			}
			end += start + len(tok.Start)

			DisplayMarkdown(w, md[:start])

			io.WriteString(w, tok.RStart)
			switch inside := md[start+len(tok.Start) : end]; tok.Type {
			default:
				DisplayMarkdown(w, inside)
			case Code, Mono:
				WriteHTML(w, inside)
			}
			io.WriteString(w, tok.REnd)

			md = md[end+len(tok.End):]
			replaced = true
			break
		}
		if !replaced {
			WriteHTML(w, md)
			break
		}
	}
}
