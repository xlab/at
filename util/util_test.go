package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytes(t *testing.T) {
	t.Parallel()

	out, err := Bytes("4160629140050E")
	exp := []byte{0x41, 0x60, 0x62, 0x91, 0x40, 0x5, 0x0e}
	assert.NoError(t, err)
	assert.Equal(t, exp, out)

	_, err = Bytes("4160629140050E0")
	assert.EqualError(t, err, "parse octets: uneven length of string")

	_, err = Bytes("4160629140050K")
	assert.EqualError(t, err, "parse octets: met a non-HEX rune in string")
}

func TestHexString(t *testing.T) {
	t.Parallel()

	buf := []byte{0x41, 0x60, 0x62, 0x91, 0x40, 0x5, 0x0e}
	out := HexString(buf)
	exp := "4160629140050E"
	assert.Equal(t, exp, out)
}
