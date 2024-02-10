package styles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMerge(t *testing.T) {
	for name, test := range map[string]struct {
		styles   []Props
		expected Props
	}{
		"2 styles": {
			styles: []Props{
				{
					Width: "100px",
					Color: "blue",
				},
				{
					Color:           "red", // This should override the blue color in baseStyle
					BackgroundColor: "yellow",
				},
			},
			expected: Props{
				Width:           "100px",
				Color:           "red",
				BackgroundColor: "yellow",
			},
		},
		"3 styles": {
			styles: []Props{
				{
					Padding: "10px",
					Margin:  "5px",
				},
				{
					Margin: "10px", // This should override the baseStyle margin
					Color:  "red",
				},
				{
					Color:  "blue", // This should override the secondaryStyle color
					Border: "1px solid black",
				},
			},
			expected: Props{
				Padding: "10px",
				Margin:  "10px", // From secondaryStyle
				Color:   "blue", // From tertiaryStyle
				Border:  "1px solid black",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, Merge(test.styles...))
		})
	}
}
