package nullable

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNullIgnoreScanner(t *testing.T) {
	{
		v := 0
		s := NewNullIgnoreScanner(&v)
		s.Scan(2)

		require.Equal(t, 2, v)
	}

	{
		v := 0
		s := NewNullIgnoreScanner(&v)
		s.Scan(nil)

		require.Equal(t, 0, v)
	}
}
