package sms

import "time"

// ValidityPeriodFormat represents the format of message's validity period.
type ValidityPeriodFormat byte

// ValidityPeriodFormats represent the possible formats of message's
// validity period (3GPP TS 23.040).
var ValidityPeriodFormats = struct {
	FieldNotPresent ValidityPeriodFormat
	Relative        ValidityPeriodFormat
	Enhanced        ValidityPeriodFormat
	Absolute        ValidityPeriodFormat
}{
	0x00, 0x02, 0x01, 0x03,
}

// Absolute validity period (3GPP TS 23.040 9.2.3.12.2)
type AbsoluteValidityPeriod = Timestamp

// Relative validity period (3GPP TS 23.040 9.2.3.12.1)
type RelativeValidityPeriod time.Duration

// Type alias for backwards compatibility
type ValidityPeriod = RelativeValidityPeriod

// Octet return a one-byte representation of the validity period.
func (v RelativeValidityPeriod) Octet() byte {
	switch d := time.Duration(v); {
	case d/time.Minute < 5:
		return 0x00
	case d/time.Hour < 12:
		return byte(d / (time.Minute * 5))
	case d/time.Hour < 24:
		return byte((d-d/time.Hour*12)/(time.Minute*30) + 143)
	case d/time.Hour < 744:
		days := d / (time.Hour * 24)
		return byte(days + 166)
	default:
		weeks := d / (time.Hour * 24 * 7)
		if weeks > 62 {
			return 0xFF
		}
		return byte(weeks + 192)
	}
}

// ReadFrom reads the validity period form the given byte.
func (v *RelativeValidityPeriod) ReadFrom(oct byte) {
	switch n := time.Duration(oct); {
	case n >= 0 && n <= 143:
		*v = RelativeValidityPeriod(5 * time.Minute * n)
	case n >= 144 && n <= 167:
		*v = RelativeValidityPeriod(12*time.Hour + 30*time.Minute*(n-143))
	case n >= 168 && n <= 196:
		*v = RelativeValidityPeriod(24 * time.Hour * (n - 166))
	case n >= 197 && n <= 255:
		*v = RelativeValidityPeriod(7 * 24 * time.Hour * (n - 192))
	}
}
