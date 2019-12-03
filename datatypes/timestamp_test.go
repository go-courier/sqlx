package datatypes

import (
	"testing"
	"time"

	"github.com/go-courier/testingx"
	"github.com/onsi/gomega"
)

func TestTimestamp(t *testing.T) {
	t.Run("Parse", testingx.It(func(t *testingx.T) {
		t0, _ := time.Parse(time.RFC3339, "2017-03-27T23:58:59+08:00")
		dt := Timestamp(t0)

		t.Expect(dt.String()).To(gomega.Equal("2017-03-27T23:58:59+08:00"))
		t.Expect(dt.Format(time.RFC3339)).To(gomega.Equal("2017-03-27T23:58:59+08:00"))
		t.Expect(dt.Unix()).To(gomega.Equal(int64(1490630339)))
	}))

	t.Run("Marshal & Unmarshal", testingx.It(func(t *testingx.T) {
		t0, _ := time.Parse(time.RFC3339, "2017-03-27T23:58:59+08:00")
		dt := Timestamp(t0)

		dateString, err := dt.MarshalText()
		t.Expect(err).To(gomega.BeNil())
		t.Expect(string(dateString)).To(gomega.Equal("2017-03-27T23:58:59+08:00"))

		dt2 := TimestampZero
		t.Expect(dt2.IsZero()).To(gomega.BeTrue())

		err = dt2.UnmarshalText(dateString)
		t.Expect(err).To(gomega.BeNil())
		t.Expect(dt2).To(gomega.Equal(dt))
		t.Expect(dt2.IsZero()).To(gomega.BeFalse())

		dt3 := TimestampZero
		err = dt3.UnmarshalText([]byte(""))
		t.Expect(err).To(gomega.BeNil())
	}))
}
