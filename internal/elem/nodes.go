package elem

import (
	"strings"

	"github.com/rprtr258/flatnotes/internal/elem/attrs"
)

// Raw takes html content and returns a RawNode.
func Raw(html string) Node {
	return func(sb *strings.Builder) {
		sb.WriteString(html)
	}
}

// None represents a node that renders nothing.
func None() Node {
	return func(*strings.Builder) {}
}

// HTML creates an <html> element.
func HTML(attrs attrs.Props, children ...Node) Node {
	el := Element("html", attrs, children...)
	return func(sb *strings.Builder) {
		sb.WriteString("<!DOCTYPE html>")
		el(sb)
	}
}

// Comment creates a CommentNode.
func Comment(comment string) Node {
	return func(sb *strings.Builder) {
		sb.WriteString("<!-- ")
		sb.WriteString(comment)
		sb.WriteString(" -->")
	}
}
