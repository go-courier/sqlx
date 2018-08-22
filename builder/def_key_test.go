package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeys(t *testing.T) {
	tt := require.New(t)

	keys := Keys{}

	tt.Equal(0, keys.Len())

	{
		keys.Add(PrimaryKey())
		tt.Equal(1, keys.Len())
	}
}
