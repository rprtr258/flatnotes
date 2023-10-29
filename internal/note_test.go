package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReImageBase64(t *testing.T) {
	text := `абоба ![image.png](data:image/png;base64,iVBO) aboba`
	got := _reImageBase64.ReplaceAllString(text, "")
	assert.Equal(t, `абоба  aboba`, got)
}
