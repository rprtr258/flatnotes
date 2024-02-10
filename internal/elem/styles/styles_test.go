package styles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToInline(t *testing.T) {
	for name, test := range map[string]struct {
		props    Props
		expected string
	}{
		"ToInline": {
			props: Props{
				BackgroundColor: "blue",
				Color:           "white",
				FontSize:        "16px",
			},
			expected: "background-color: blue;color: white;font-size: 16px;",
		},
		"Empty": {
			props:    Props{},
			expected: "",
		},
		"SinglePair": {
			props: Props{
				Color: "red",
			},
			expected: "color: red;",
		},
		"UnorderedKeys": {
			props: Props{
				BackgroundColor: "blue",
				Color:           "white",
				FontSize:        "16px",
				OutlineStyle:    "solid",
			},
			expected: "background-color: blue;color: white;font-size: 16px;outline-style: solid;",
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.props.ToInline())
		})
	}
}
