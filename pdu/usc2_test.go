package pdu

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testStringUcs2 = "Этот абонент звонил вам 2 раза"

var testOctetsUcs2 = []byte{
	0x04, 0x2D, 0x04, 0x42, 0x04, 0x3E, 0x04, 0x42,
	0x00, 0x20, 0x04, 0x30, 0x04, 0x31, 0x04, 0x3E,
	0x04, 0x3D, 0x04, 0x35, 0x04, 0x3D, 0x04, 0x42,
	0x00, 0x20, 0x04, 0x37, 0x04, 0x32, 0x04, 0x3E,
	0x04, 0x3D, 0x04, 0x38, 0x04, 0x3B, 0x00, 0x20,
	0x04, 0x32, 0x04, 0x30, 0x04, 0x3C, 0x00, 0x20,
	0x00, 0x32, 0x00, 0x20, 0x04, 0x40, 0x04, 0x30,
	0x04, 0x37, 0x04, 0x30,
}

func TestEncodeUcs2(t *testing.T) {
	t.Parallel()

	out := EncodeUcs2(testStringUcs2)
	exp := testOctetsUcs2
	assert.Equal(t, exp, out)
}

func TestDecodeUcs2(t *testing.T) {
	t.Parallel()

	oct := testOctetsUcs2
	out, err := DecodeUcs2(oct, false)
	exp := testStringUcs2
	require.NoError(t, err)
	assert.Equal(t, exp, out)
}
