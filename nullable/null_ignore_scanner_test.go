package nullable

import (
	"testing"

	"github.com/go-courier/testingx"
	"github.com/onsi/gomega"
)

func TestNullIgnoreScanner(t *testing.T) {
	t.Run("scan value", testingx.It(func(t *testingx.T) {
		v := 0
		s := NewNullIgnoreScanner(&v)
		s.Scan(2)

		t.Expect(v).To(gomega.Equal(2))
	}))

	t.Run("scan nil", testingx.It(func(t *testingx.T) {
		v := 0
		s := NewNullIgnoreScanner(&v)
		s.Scan(nil)

		t.Expect(v).To(gomega.Equal(0))
	}))
}
