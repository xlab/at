package pdu

import (
	"errors"
	"unicode/utf16"
)

// ErrUnevenNumber happens when the number of octets (bytes) in the input is uneven.
var ErrUnevenNumber = errors.New("decode ucs2: uneven number of octets")

// EncodeUcs2 encodes the given UTF-8 text into UCS2 (UTF-16) encoding and returns the produced octets.
func EncodeUcs2(str string) []byte {
	buf := utf16.Encode([]rune(str))
	octets := make([]byte, 0, len(buf)*2)
	for _, n := range buf {
		octets = append(octets, byte(n&0xFF00>>8), byte(n&0x00FF))
	}
	return octets
}

// DecodeUcs2 decodes the given UCS2 (UTF-16) octet data into a UTF-8 encoded string.
func DecodeUcs2(octets []byte, startsWithHeader bool) (str string, err error) {
	octetsLng := len(octets)
	headerLng := 0

	if octetsLng == 0 {
		err = ErrIncorrectDataLength
		return
	}

	if startsWithHeader {
		// just ignore header
		headerLng = int(octets[0]) + 1
		if (octetsLng - headerLng) <= 0 {
			err = ErrIncorrectDataLength
			return
		}

		octetsLng = octetsLng - headerLng
	}

	if octetsLng%2 != 0 {
		err = ErrUnevenNumber
		return
	}
	buf := make([]uint16, 0, octetsLng/2)
	for i := 0; i < octetsLng; i += 2 {
		buf = append(buf, uint16(octets[i+headerLng])<<8|uint16(octets[i+1+headerLng]))
	}
	runes := utf16.Decode(buf)
	return string(runes), nil
}
