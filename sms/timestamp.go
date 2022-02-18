package sms

import (
	"time"

	"github.com/xlab/at/pdu"
)

// Timestamp represents message's timestamp.
type Timestamp time.Time

// PDU returns bytes of semi-octet encoded timestamp, as specified in
// 3GPP TS 23.040 version 16.0.0 release 16, section 9.2.3.11.
//
// TP-Service-Centre-Time-Stamp (TP-SCTS)
//
//  |             | Year | Month | Day | Hour | Minute | Second | Time Zone |
//  |-------------|------|-------|-----|------|--------|--------|-----------|
//  | Semi-octets |   2  |   2   |  2  |   2  |    2   |    2   |     2     |
//
// The Time Zone indicates the difference, expressed in quarters of an hour,
// between the local time and GMT. In the first of the two semi-octets, the
// first bit (bit 3 of the seventh octet of the TP-Service-CentreTime-Stamp
// field) represents the algebraic sign of this difference (0: positive,
// 1: negative).
func (t Timestamp) PDU() []byte {
	date := time.Time(t)
	year, month, day := date.Date()
	hour, minute, second := date.Clock()

	_, offset := date.Zone()
	negativeOffset := offset < 0 //nolint:ifshort // false positive
	if negativeOffset {
		offset = -offset
	}
	quarters := offset / int(time.Hour/time.Second) * 4

	octets := []byte{
		/* YY */ pdu.Swap(pdu.Encode((year % 1000))),
		/* MM */ pdu.Swap(pdu.Encode(int(month))),
		/* DD */ pdu.Swap(pdu.Encode(day)),
		/* hh */ pdu.Swap(pdu.Encode(hour)),
		/* mm */ pdu.Swap(pdu.Encode(minute)),
		/* ss */ pdu.Swap(pdu.Encode(second)),
		/* zz */ pdu.Swap(pdu.Encode(quarters)),
	}
	if negativeOffset {
		octets[6] |= 0x04
	}

	return octets
}

// ReadFrom reads a semi-encoded timestamp from the given octets.
// See (*Timestamp).PDU() for format details.
func (t *Timestamp) ReadFrom(octets []byte) {
	millennium := (time.Now().Year() / 1000) * 1000
	year := pdu.Decode(pdu.Swap(octets[0]))
	month := pdu.Decode(pdu.Swap(octets[1]))
	day := pdu.Decode(pdu.Swap(octets[2]))
	hour := pdu.Decode(pdu.Swap(octets[3]))
	minute := pdu.Decode(pdu.Swap(octets[4]))
	second := pdu.Decode(pdu.Swap(octets[5]))

	negativeOffset := (octets[6] & 0x04) != 0
	quarters := pdu.Decode(pdu.Swap(octets[6] & 0xF7))
	offset := time.Duration(quarters) * 15 * time.Minute

	date := time.Date(millennium+year, time.Month(month), day, hour, minute, second, 0, time.UTC)

	if negativeOffset {
		offset = -offset
	}
	date = date.Add(-offset).In(time.FixedZone("", int(offset.Seconds())))
	*t = Timestamp(date)
}
