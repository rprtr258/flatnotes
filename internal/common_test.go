package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func use[T any](t T, err error) func(*testing.T) T {
	if err != nil {
		return func(t *testing.T) T {
			require.NoError(t, err)
			panic(err)
		}
	}
	return func(*testing.T) T {
		return t
	}
}
