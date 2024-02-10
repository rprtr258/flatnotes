package elem

import "strings"

// None represents a node that renders nothing.
func None() Node {
	return func(*strings.Builder) {}
}
