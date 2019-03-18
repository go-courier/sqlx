package generator

import (
	"os"
	"testing"

	"github.com/go-courier/sqlx/v2/builder"
	"github.com/stretchr/testify/require"
)

func init() {
	os.Chdir("./test")
}

func TestParseIndexesFromDoc(t *testing.T) {
	tt := require.New(t)

	{
		keys, other := parseKeysFromDoc(`
@def primary ID
summary
desc
`)

		tt.Equal(&Keys{
			Primary: []string{"ID"},
		}, keys)

		tt.Equal([]string{
			"summary",
			"desc",
		}, other)

	}
	{
		keys, _ := parseKeysFromDoc(`
@def index I_name   Name
@def index I_nickname/HASH Nickname Name
	`)

		tt.Equal(&Keys{
			Indexes: builder.Indexes{
				"I_name":          []string{"Name"},
				"I_nickname/HASH": []string{"Nickname", "Name"},
			},
		}, keys)
	}

	{
		keys, _ := parseKeysFromDoc(`
@def primary ID
@def index I_nickname/BTREE Nickname
@def index I_username Username
@def index I_geom/SPATIAL Geom
@def unique_index I_name Name
	`)

		tt.Equal(&Keys{
			Primary: []string{"ID"},
			Indexes: builder.Indexes{
				"I_nickname/BTREE": []string{"Nickname"},
				"I_username":       []string{"Username"},
				"I_geom/SPATIAL":   []string{"Geom"},
			},
			UniqueIndexes: builder.Indexes{
				"I_name": []string{"Name"},
			},
		}, keys)
	}
}

func TestParseColRel(t *testing.T) {
	rel, others := parseColRelFromComment(`
@rel Account.AccountID

summary

desc
`)
	require.Equal(t, rel, "Account.AccountID")
	require.Equal(t, []string{
		"summary",
		"desc",
	}, others)
}
