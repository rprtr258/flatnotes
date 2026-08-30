package styles

import (
	"sort"
	"strings"

	"github.com/rprtr258/fun"
)

// Props is a map of CSS properties
type Props map[string]string

// ToInline converts the Props to an inline style string
func (p Props) ToInline() string {
	// Extract the keys and sort them for deterministic order
	keys := fun.Keys(p)
	sort.Strings(keys)

	var sb strings.Builder
	for _, key := range keys {
		sb.WriteString(key)
		sb.WriteString(": ")
		sb.WriteString(p[key])
		sb.WriteString(";")
	}
	return sb.String()
}
