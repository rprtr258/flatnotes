package elem

import (
	"sort"
	"strings"

	"github.com/rprtr258/fun"
	"github.com/rprtr258/fun/set"

	"github.com/rprtr258/flatnotes/internal/elem/attrs"
)

// List of HTML5 void elements. Void elements, also known as self-closing or empty elements,
// are elements that don't have a closing tag because they can't contain any content.
// For example, the <img> tag cannot wrap text or other tags, it stands alone, so it doesn't have a closing tag.
var voidElements = set.NewFrom(
	"area",
	"base",
	"br",
	"col",
	"command",
	"embed",
	"hr",
	"img",
	"input",
	"keygen",
	"link",
	"meta",
	"param",
	"source",
	"track",
	"wbr",
)

// List of boolean attributes. Boolean attributes can't have literal values. The presence of an boolean
// attribute represents the "true" value. To represent the "false" value, the attribute has to be omitted.
// See https://html.spec.whatwg.org/multipage/indices.html#attributes-3 for reference
var booleanAttrs = set.NewFrom(
	attrs.AllowFullscreen,
	attrs.Async,
	attrs.Autofocus,
	attrs.Autoplay,
	attrs.Checked,
	attrs.Controls,
	attrs.Defer,
	attrs.Disabled,
	attrs.Ismap,
	attrs.Loop,
	attrs.Multiple,
	attrs.Muted,
	attrs.Novalidate,
	attrs.Open,
	attrs.Playsinline,
	attrs.Readonly,
	attrs.Required,
	attrs.Selected,
)

func Element(tag string, attrs attrs.Props, children ...Node) Node {
	return func(sb *strings.Builder) {
		// Start with opening tag
		sb.WriteString("<")
		sb.WriteString(tag)

		// Sort the keys for consistent order
		keys := fun.Keys(attrs)
		sort.Strings(keys)

		// Append the attributes to the builder
		for _, attrName := range keys {
			if !booleanAttrs.Contains(attrName) {
				// regular attribute has a name and a value
				sb.WriteString(` `)
				sb.WriteString(attrName)
				sb.WriteString(`="`)
				sb.WriteString(attrs[attrName])
				sb.WriteString(`"`)
			} else if attrs[attrName] == "true" {
				// boolean attribute presents its name only if the value is "true"
				sb.WriteString(` `)
				sb.WriteString(attrName)
			}
		}

		// If it's a void element, close it and return
		if voidElements.Contains(tag) {
			sb.WriteString(`>`)
			return
		}

		// Close opening tag
		sb.WriteString(`>`)

		// Build the content
		for _, child := range children {
			child(sb)
		}

		// Append closing tag
		sb.WriteString(`</`)
		sb.WriteString(tag)
		sb.WriteString(`>`)
	}

}
