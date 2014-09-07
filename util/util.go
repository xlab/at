// Package util provides some useful methods that are used within the 'at' package. These methods include a method to
// extract bytes from a string.
package util

import (
	"errors"
	"fmt"
	"strconv"
)

// Common errors.
var (
	ErrUnevenLength = errors.New("parse octets: uneven length of string")
	ErrUnexpected   = errors.New("parse octets: met a non-HEX rune in string")
)

// Bytes parses the hex-string of odd length into bytes.
func Bytes(hex string) ([]byte, error) {
	if len(hex)%2 != 0 {
		return nil, ErrUnevenLength
	}
	octets := make([]byte, 0, len(hex)/2)
	for i := 0; i < len(hex); i += 2 {
		frame := hex[i : i+2]
		oct, err := strconv.ParseUint(frame, 16, 8)
		if err != nil {
			return nil, ErrUnexpected
		}
		octets = append(octets, byte(oct))
	}
	return octets, nil
}

// MustBytes is an alias for Bytes, except that it will panic
// if there is any parse error.
func MustBytes(hex string) []byte {
	b, err := Bytes(hex)
	if err != nil {
		panic(err)
	}
	return b
}

// HexString produces a hex-string from bytes. Like a DEADBEEF, without prepending the 0x.
func HexString(octets []byte) string {
	return fmt.Sprintf("%2X", octets)
}
