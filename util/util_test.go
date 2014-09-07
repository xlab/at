package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytes(t *testing.T) {
	out, err := Bytes("4160629140050E")
	_, err2 := Bytes("4160629140050E0")
	_, err3 := Bytes("4160629140050K")
	exp := []byte{0x41, 0x60, 0x62, 0x91, 0x40, 0x5, 0x0e}
	assert.NoError(t, err)
	assert.Error(t, err2)
	assert.Error(t, err3)
	assert.Equal(t, exp, out)
}

func TestHexString(t *testing.T) {
	buf := []byte{0x41, 0x60, 0x62, 0x91, 0x40, 0x5, 0x0e}
	out := HexString(buf)
	exp := "4160629140050E"
	assert.Equal(t, exp, out)
}
