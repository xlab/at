package sms

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/xlab/at/pdu"
)

// PhoneNumber represents the address in either local or international format.
type PhoneNumber string

// PDU returns the number of digits in address and octets of semi-octet encoded address.
func (p PhoneNumber) PDU() (int, []byte, error) {
	digitStr := strings.TrimPrefix(string(p), "+")
	var str string
	for _, r := range digitStr {
		if r >= '0' && r <= '9' {
			str = str + string(r)
		}
	}
	n := len(str)
	number, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(p.Type())
	buf.Write(pdu.EncodeSemi(number))
	return n, buf.Bytes(), nil
}

// Type returns the type of address — local or international.
func (p PhoneNumber) Type() byte {
	if strings.HasPrefix(string(p), "+") {
		return (0x81 | 0x10) // 1001 0001
	}
	return (0x81 | 0x20) // 1010 0001
}

// ReadFrom constructs an address from the semi-decoded version in the supplied byte slice.
func (p *PhoneNumber) ReadFrom(octets []byte) {
	if len(octets) < 1 {
		return
	}
	addrType := octets[0]

	// Alphanumeric, (coded according to GSM TS 03.38 7-bit default alphabet)
	if addrType&0x70 == 0x50 {
		// decode 7 bit
		addr, err := pdu.Decode7Bit(octets[1:])
		if err != nil {
			// handle error, panic or log.Println("Decode7bit", octets, err)
		}
		*p = PhoneNumber(addr)
		return
	}

	addr := pdu.DecodeSemiAddress(octets[1:])
	if addrType&0x10 > 0 {
		*p = PhoneNumber("+" + addr)
	} else {
		*p = PhoneNumber(addr)
	}
	return
}
