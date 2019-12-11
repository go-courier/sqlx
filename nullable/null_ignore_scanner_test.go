package nullable

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestNullIgnoreScanner(t *testing.T) {
	t.Run("scan value", func(t *testing.T) {
		v := 0
		s := NewNullIgnoreScanner(&v)
		s.Scan(2)

		gomega.NewWithT(t).Expect(v).To(gomega.Equal(2))
	})

	t.Run("scan nil", func(t *testing.T) {
		v := 0
		s := NewNullIgnoreScanner(&v)
		s.Scan(nil)

		gomega.NewWithT(t).Expect(v).To(gomega.Equal(0))
	})
}
