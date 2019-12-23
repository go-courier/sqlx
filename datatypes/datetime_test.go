package datatypes

import (
	"testing"
	"time"

	"github.com/onsi/gomega"
)

func TestDatetime(t *testing.T) {
	t.Run("Parse", func(t *testing.T) {
		t0, _ := time.Parse(time.RFC3339, "2017-03-27T23:58:59+08:00")
		dt := Datetime(t0)

		gomega.NewWithT(t).Expect(dt.String()).To(gomega.Equal("2017-03-27T23:58:59+08:00"))
		gomega.NewWithT(t).Expect(dt.Format(time.RFC3339)).To(gomega.Equal("2017-03-27T23:58:59+08:00"))
		gomega.NewWithT(t).Expect(dt.Unix()).To(gomega.Equal(int64(1490630339)))
	})
	t.Run("Marshal & Unmarshal", func(t *testing.T) {
		t0, _ := time.Parse(time.RFC3339, "2017-03-27T23:58:59+08:00")
		dt := Datetime(t0)

		dateString, err := dt.MarshalText()
		gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
		gomega.NewWithT(t).Expect(string(dateString)).To(gomega.Equal("2017-03-27T23:58:59+08:00"))

		dt2 := DatetimeZero
		gomega.NewWithT(t).Expect(dt2.IsZero()).To(gomega.BeTrue())

		err = dt2.UnmarshalText(dateString)
		gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
		gomega.NewWithT(t).Expect(dt2).To(gomega.Equal(dt))
		gomega.NewWithT(t).Expect(dt2.IsZero()).To(gomega.BeFalse())

		dt3 := TimestampZero
		err = dt3.UnmarshalText([]byte(""))
		gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
	})

	t.Run("Scan & Value", func(t *testing.T) {
		t0, _ := time.Parse(time.RFC3339, "2017-03-27T23:58:59+08:00")
		dt := Datetime(t0)

		value, err := dt.Value()
		gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
		gomega.NewWithT(t).Expect(value.(time.Time).In(CST).Format(time.RFC3339)).To(gomega.Equal("2017-03-27T23:58:59+08:00"))

		dt2 := DatetimeZero
		gomega.NewWithT(t).Expect(dt2.IsZero()).To(gomega.BeTrue())

		err = dt2.Scan(value)
		gomega.NewWithT(t).Expect(err).To(gomega.BeNil())
		gomega.NewWithT(t).Expect(dt2.In(CST)).To(gomega.Equal(dt.In(CST)))
		gomega.NewWithT(t).Expect(dt2.IsZero()).To(gomega.BeFalse())
	})
}
