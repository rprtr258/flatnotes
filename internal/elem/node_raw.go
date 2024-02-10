package elem

import "strings"

// Raw takes html content and returns a RawNode.
func Raw(html string) Node {
	return func(sb *strings.Builder) {
		sb.WriteString(html)
	}
}
