package fts

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestTokenizer(t *testing.T) {
// 	for _, test := range []struct {
// 		text   string
// 		tokens []string
// 	}{
// 		{
// 			text:   "",
// 			tokens: []string{},
// 		},
// 		{
// 			text:   "a",
// 			tokens: []string{"a"},
// 		},
// 		{
// 			text:   "small wild,cat!",
// 			tokens: []string{"small", "wild", "cat"},
// 		},
// 	} {
// 		t.Run(test.text, func(st *testing.T) {
// 			assert.EqualValues(st, test.tokens, tokenize(test.text))
// 		})
// 	}
// }
