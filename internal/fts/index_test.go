package fts

// import (
// 	"testing"

// 	"github.com/stretchr/testify/require"
// )

// func TestIndex(t *testing.T) {
// 	idx := NewIndex[document]()

// 	require.Nil(t, idx.Search("foo"))
// 	require.Nil(t, idx.Search("donut"))

// 	idx.Add(document{Id: "1", Text: "A donut on a glass plate. Only the donuts."})
// 	require.Nil(t, idx.Search("a"))
// 	require.Equal(t, idx.Search("donut"), []int{1})
// 	require.Equal(t, idx.Search("DoNuts"), []int{1})
// 	require.Equal(t, idx.Search("glass"), []int{1})

// 	idx.Add(document{Id: "2", Text: "donut is a donut"})
// 	require.Nil(t, idx.Search("a"))
// 	require.Equal(t, idx.Search("donut"), []int{1, 2})
// 	require.Equal(t, idx.Search("DoNuts"), []int{1, 2})
// 	require.Equal(t, idx.Search("glass"), []int{1})
// }
