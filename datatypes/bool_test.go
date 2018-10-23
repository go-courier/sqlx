package datatypes

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBool(t *testing.T) {
	tt := require.New(t)
	{
		bytes, _ := json.Marshal(BOOL_TRUE)
		tt.Equal("true", string(bytes))
	}
	{
		bytes, _ := json.Marshal(BOOL_FALSE)
		tt.Equal("false", string(bytes))
	}
	{
		bytes, _ := json.Marshal(BOOL_UNKNOWN)
		tt.Equal("null", string(bytes))
	}

	{
		var b Bool
		json.Unmarshal([]byte("true"), &b)
		tt.Equal(BOOL_TRUE, b)
	}
	{
		var b Bool
		json.Unmarshal([]byte("false"), &b)
		tt.Equal(BOOL_FALSE, b)
	}
	{
		var b Bool
		json.Unmarshal([]byte("null"), &b)
		tt.Equal(BOOL_UNKNOWN, b)
	}
}
