package pdu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeSemi(t *testing.T) {
	out := EncodeSemi(14, 06, 26, 21, 36, 30, 16)
	exp := []byte{0x41, 0x60, 0x62, 0x12, 0x63, 0x03, 0x61}
	assert.Equal(t, exp, out)
}

func TestDecodeSemi(t *testing.T) {
	oct := []byte{0x41, 0x60, 0x62, 0x12, 0x63, 0x03, 0x61}
	out := DecodeSemi(oct)
	exp := []int{14, 06, 26, 21, 36, 30, 16}
	assert.Equal(t, exp, out)
}
