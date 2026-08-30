package fts

// import (
// 	"testing"

// 	"github.com/stretchr/testify/require"
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
// 			require.EqualValues(st, test.tokens, tokenize(test.text))
// 		})
// 	}
// }
