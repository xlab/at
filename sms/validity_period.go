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

// ValidityPeriod represents the validity period of message.
type ValidityPeriod time.Duration

// Octet return a one-byte representation of the validity period.
func (v ValidityPeriod) Octet() byte {
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
func (v *ValidityPeriod) ReadFrom(oct byte) {
	switch n := time.Duration(oct); {
	case n >= 0 && n <= 143:
		*v = ValidityPeriod(5 * time.Minute * n)
	case n >= 144 && n <= 167:
		*v = ValidityPeriod(12*time.Hour + 30*time.Minute*(n-143))
	case n >= 168 && n <= 196:
		*v = ValidityPeriod(24 * time.Hour * (n - 166))
	case n >= 197 && n <= 255:
		*v = ValidityPeriod(7 * 24 * time.Hour * (n - 192))
	}
}
