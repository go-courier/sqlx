package generator

import (
	"os"
	"testing"

	"github.com/go-courier/sqlx/builder"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Chdir("./test")
}

func TestParseIndexesFromDoc(t *testing.T) {
	tt := assert.New(t)

	tt.Equal(&Keys{
		Primary: []string{"ID"},
	}, parseKeysFromDoc(`
	@def primary ID
	`))

	tt.Equal(&Keys{
		Indexes: builder.Indexes{
			"I_name":     []string{"Name"},
			"I_nickname": []string{"Nickname", "Name"},
		},
	}, parseKeysFromDoc(`
	@def index I_name   Name
	@def index I_nickname   Nickname Name
	`))

	tt.Equal(&Keys{
		Primary: []string{"ID"},
		Indexes: builder.Indexes{
			"I_nickname": []string{"Nickname", "Name"},
		},
		UniqueIndexes: builder.Indexes{
			"I_name": []string{"Name"},
		},
	}, parseKeysFromDoc(`
	@def primary ID
	@def index I_nickname Nickname Name
	@def unique_index I_name Name
	`))
}
