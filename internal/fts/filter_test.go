package fts

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestFilters(t *testing.T) {
// 	for name, test := range map[string]struct {
// 		in, out []string
// 		filter  func([]string) []string
// 	}{
// 		"lowercase": {
// 			in:     []string{"Cat", "DOG", "fish"},
// 			out:    []string{"cat", "dog", "fish"},
// 			filter: lowercaseFilter,
// 		},
// 		"stopword": {
// 			in:     []string{"i", "am", "the", "cat"},
// 			out:    []string{"am", "cat"},
// 			filter: stopwordFilter,
// 		},
// 		"stemmer": {
// 			in:     []string{"cat", "cats", "fish", "fishing", "fished", "airline"},
// 			out:    []string{"cat", "cat", "fish", "fish", "fish", "airlin"},
// 			filter: stemmerFilter,
// 		},
// 	} {
// 		t.Run(name, func(t *testing.T) {
// 			assert.Equal(t, test.out, test.filter(test.in))
// 		})
// 	}
// }
