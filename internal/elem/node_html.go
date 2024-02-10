package elem

import (
	"strings"

	"github.com/rprtr258/flatnotes/internal/elem/attrs"
)

// HTML creates an <html> element.
func HTML(attrs attrs.Props, children ...Node) Node {
	el := Element("html", attrs, children...)
	return func(sb *strings.Builder) {
		sb.WriteString("<!DOCTYPE html>")
		el(sb)
	}
}
