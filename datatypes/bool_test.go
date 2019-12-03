package datatypes

import (
	"encoding/json"
	"testing"

	"github.com/go-courier/testingx"
	"github.com/onsi/gomega"
)

func TestBool(t *testing.T) {
	t.Run("Marshal", testingx.It(func(t *testingx.T) {
		bytes, _ := json.Marshal(BOOL_TRUE)
		t.Expect(string(bytes)).To(gomega.Equal("true"))

		bytes, _ = json.Marshal(BOOL_FALSE)
		t.Expect(string(bytes)).To(gomega.Equal("false"))

		bytes, _ = json.Marshal(BOOL_UNKNOWN)
		t.Expect(string(bytes)).To(gomega.Equal("null"))
	}))

	t.Run("Unmarshal", testingx.It(func(t *testingx.T) {
		var b Bool

		json.Unmarshal([]byte("null"), &b)
		t.Expect(b).To(gomega.Equal(BOOL_UNKNOWN))

		json.Unmarshal([]byte("true"), &b)
		t.Expect(b).To(gomega.Equal(BOOL_TRUE))

		json.Unmarshal([]byte("false"), &b)
		t.Expect(b).To(gomega.Equal(BOOL_FALSE))
	}))
}
