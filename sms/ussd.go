package sms

import "github.com/xlab/at/pdu"

// USSD represents an USSD query string.
type USSD string

// Gsm7Bit encodes USSD query into GSM 7-Bit packed octets.
func (u USSD) Gsm7Bit() []byte {
	return pdu.Encode7Bit(string(u))
}
