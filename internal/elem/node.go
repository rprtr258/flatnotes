package elem

import "strings"

type Node func(*strings.Builder)

func Render(n Node) string {
	var sb strings.Builder
	n(&sb)
	return sb.String()
}
