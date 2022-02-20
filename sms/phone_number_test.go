package sms

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xlab/at/util"
)

func TestPhoneNumber(t *testing.T) {
	t.Parallel()

	type testcase struct {
		pdu    []byte
		number string
		typ    PhoneNumberType
	}

	for name, tc := range map[string]testcase{
		"international": {
			pdu:    util.MustBytes("9121436587F9"),
			number: "+123456789",
			typ:    PhoneNumberTypes.International,
		},
		"national": {
			pdu:    util.MustBytes("A11032547698"),
			number: "0123456789",
			typ:    PhoneNumberTypes.National,
		},
		"alphanumeric": {
			pdu:    util.MustBytes("D061F1985C3603"),
			number: "abcdef",
			// FIXME: we don't have proper support for alphanumeric numbers
			// yet, so Type() will just use "national" as type.
			typ: PhoneNumberTypes.National,
		},
	} {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var subject PhoneNumber
			err := subject.ReadFrom(tc.pdu)
			require.NoError(t, err)

			assert.EqualValues(t, tc.number, subject)
			assert.Equal(t, 0x81|byte(tc.typ), subject.Type())
		})
	}
}
