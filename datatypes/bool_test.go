package datatypes

import (
	"encoding/json"
	"testing"

	"github.com/onsi/gomega"
)

func TestBool(t *testing.T) {
	t.Run("Marshal", func(t *testing.T) {
		bytes, _ := json.Marshal(BOOL_TRUE)
		gomega.NewWithT(t).Expect(string(bytes)).To(gomega.Equal("true"))

		bytes, _ = json.Marshal(BOOL_FALSE)
		gomega.NewWithT(t).Expect(string(bytes)).To(gomega.Equal("false"))

		bytes, _ = json.Marshal(BOOL_UNKNOWN)
		gomega.NewWithT(t).Expect(string(bytes)).To(gomega.Equal("null"))
	})
	t.Run("Unmarshal", func(t *testing.T) {
		var b Bool

		json.Unmarshal([]byte("null"), &b)
		gomega.NewWithT(t).Expect(b).To(gomega.Equal(BOOL_UNKNOWN))

		json.Unmarshal([]byte("true"), &b)
		gomega.NewWithT(t).Expect(b).To(gomega.Equal(BOOL_TRUE))

		json.Unmarshal([]byte("false"), &b)
		gomega.NewWithT(t).Expect(b).To(gomega.Equal(BOOL_FALSE))
	})
}
