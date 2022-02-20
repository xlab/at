package sms

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xlab/at/util"
)

func TestTimestamp_PDU(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		date     string
		expected string
	}{
		{"2021-03-04T05:06:07+08:15", "12304050607033"},
		{"2021-03-04T05:06:07-08:15", "1230405060703B"},
		{"2000-01-01T00:00:00Z", "00101000000000"},
		{"1999-12-31T23:59:59Z", "99211332959500"},
	} {
		ts := parseTimestamp(tc.date)
		actual := util.HexString(ts.PDU())
		assert.Equal(t, tc.expected, actual)
	}
}

func TestTimestamp_ReadFrom(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		pdu      string
		expected string
	}{
		{"12304050607023", "2021-03-04T05:06:07+08:00"},
		{"12304050607033", "2021-03-04T05:06:07+08:15"},
		{"1230405060703B", "2021-03-04T05:06:07-08:15"},
		{"00101000000000", "2000-01-01T00:00:00Z"},
		{"99211332959500", "2099-12-31T23:59:59Z"}, // [sic]
	} {
		var subject Timestamp
		pdu := util.MustBytes(tc.pdu)
		subject.ReadFrom(pdu)
		assert.Equal(t, tc.expected, time.Time(subject).Format(time.RFC3339))
	}
}
