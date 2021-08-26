package nullable

import (
	"testing"

	"github.com/onsi/gomega"
)

func BenchmarkNewNullIgnoreScanner(b *testing.B) {
	v := 0
	for i := 0; i < b.N; i++ {
		_ = NewNullIgnoreScanner(&v).Scan(2)
	}
	b.Log(v)
}

func TestNullIgnoreScanner(t *testing.T) {
	t.Run("scan value", func(t *testing.T) {
		v := 0
		s := NewNullIgnoreScanner(&v)
		_ = s.Scan(2)

		gomega.NewWithT(t).Expect(v).To(gomega.Equal(2))
	})

	t.Run("scan nil", func(t *testing.T) {
		v := 0
		s := NewNullIgnoreScanner(&v)
		_ = s.Scan(nil)

		gomega.NewWithT(t).Expect(v).To(gomega.Equal(0))
	})
}
