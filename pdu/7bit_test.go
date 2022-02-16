package pdu

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xlab/at/util"
)

func TestEncode7Bit(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		str string
		exp []byte
	}{
		{"hello[world]! ы?", util.MustBytes("E8329BFDDEF0EE6F399BBCF18540BF1F")},
		{"AAAAAAAAAAAAAAB\r", util.MustBytes("C16030180C0683C16030180C0A1B0D")},
		{"AAAAAAAAAAAAAAB", util.MustBytes("C16030180C0683C16030180C0A1B")},
		{"height of eifel", util.MustBytes("E872FA8CA683DE6650396D2EB31B")},
	}
	for _, tc := range testcases {
		assert.Equal(t, tc.exp, Encode7Bit(tc.str))
	}
}

func TestDecode7Bit(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		exp   string
		pack7 []byte
	}{
		// ы -> ?
		{"hello[world]! ??", util.MustBytes("E8329BFDDEF0EE6F399BBCF18540BF1F")},
		{"AAAAAAAAAAAAAAB\r", util.MustBytes("C16030180C0683C16030180C0A1B0D")},
		{"AAAAAAAAAAAAAAB", util.MustBytes("C16030180C0683C16030180C0A1B")},
		{"height of eifel", util.MustBytes("E872FA8CA683DE6650396D2EB31B")},
	}
	for _, tc := range testcases {
		log.Println(displayPack(tc.pack7))
		out, err := Decode7Bit(tc.pack7)
		assert.NoError(t, err)
		assert.Equal(t, tc.exp, out)
	}
}

func TestPack7Bit(t *testing.T) {
	t.Parallel()

	raw7 := []byte{Esc, 0x3c, Esc, 0x3e}
	exp := []byte{0x1b, 0xde, 0xc6, 0x7}
	assert.Equal(t, exp, pack7Bit(raw7))
}

func TestUnpack7Bit(t *testing.T) {
	t.Parallel()

	pack7 := []byte{0x1b, 0xde, 0xc6, 0x7}
	exp := []byte{Esc, 0x3c, Esc, 0x3e}
	assert.Equal(t, exp, unpack7Bit(pack7))
}
