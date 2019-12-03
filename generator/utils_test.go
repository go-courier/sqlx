package generator

import (
	"os"
	"testing"

	"github.com/go-courier/sqlx/v2/builder"
	"github.com/go-courier/testingx"
	"github.com/onsi/gomega"
)

func init() {
	os.Chdir("./test")
}

func TestParseIndexesFromDoc(t *testing.T) {
	t.Run("parse primary", testingx.It(func(t *testingx.T) {
		keys, other := parseKeysFromDoc(`
@def primary ID
summary
desc
`)
		t.Expect(keys).To(gomega.Equal(&Keys{
			Primary: []string{"ID"},
		}))

		t.Expect(other).To(gomega.Equal([]string{
			"summary",
			"desc",
		}))
	}))

	t.Run("parse index", testingx.It(func(t *testingx.T) {
		keys, _ := parseKeysFromDoc(`
@def index I_name   Name
@def index I_nickname/HASH Nickname Name
`)
		t.Expect(keys).To(gomega.Equal(&Keys{
			Indexes: builder.Indexes{
				"I_name":          []string{"Name"},
				"I_nickname/HASH": []string{"Nickname", "Name"},
			},
		}))
	}))

	t.Run("parse all", testingx.It(func(t *testingx.T) {
		keys, _ := parseKeysFromDoc(`
@def primary ID
@def index I_nickname/BTREE Nickname
@def index I_username Username
@def index I_geom/SPATIAL Geom
@def unique_index I_name Name
`)
		t.Expect(keys).To(gomega.Equal(&Keys{
			Primary: []string{"ID"},
			Indexes: builder.Indexes{
				"I_nickname/BTREE": []string{"Nickname"},
				"I_username":       []string{"Username"},
				"I_geom/SPATIAL":   []string{"Geom"},
			},
			UniqueIndexes: builder.Indexes{
				"I_name": []string{"Name"},
			},
		}))
	}))

}

func TestParseColRel(t *testing.T) {
	t.Run("rel", testingx.It(func(t *testingx.T) {
		rel, others := parseColRelFromComment(`
@rel Account.AccountID

summary

desc
`)
		t.Expect(rel).To(gomega.Equal("Account.AccountID"))
		t.Expect(others).To(gomega.Equal([]string{
			"summary",
			"desc",
		}))
	}))
}
