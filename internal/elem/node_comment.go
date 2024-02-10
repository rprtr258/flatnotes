package elem

import "strings"

// Comment creates a CommentNode.
func Comment(comment string) Node {
	return func(sb *strings.Builder) {
		sb.WriteString("<!-- ")
		sb.WriteString(comment)
		sb.WriteString(" -->")
	}
}
