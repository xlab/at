package sms

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/xlab/at/pdu"
)

// PhoneNumber represents the address in either local or international format.
type PhoneNumber string

// PhoneNumberType represents Type-of-Number, as specified in 3GPP
// TS 23.040 version 16.0.0 release 16, section 9.1.2.5.
type PhoneNumberType byte

// PhoneNumberTypes are all known PhoneNumberType values.
var PhoneNumberTypes = struct {
	Unknown         PhoneNumberType
	International   PhoneNumberType
	National        PhoneNumberType
	NetworkSpecific PhoneNumberType
	Subscriber      PhoneNumberType
	Alphanumeric    PhoneNumberType
	Abbreviated     PhoneNumberType
	Reserved        PhoneNumberType // for future extension
}{
	Unknown:         0 << 4,
	International:   1 << 4,
	National:        2 << 4,
	NetworkSpecific: 3 << 4,
	Subscriber:      4 << 4,
	Alphanumeric:    5 << 4,
	Abbreviated:     6 << 4,
	Reserved:        7 << 4,
}

// NumberingPlan represents Numbering-plan-identification, as specified
// in 3GPP TS 23.040 version 16.0.0 release 16, section 9.1.2.5.
type NumberingPlan byte

// NumberingPlans are all known NumberingPlan valus. Other values are
// reserved.
var NumberingPlans = struct {
	Unknown                NumberingPlan
	E164                   NumberingPlan // ISDN/telephone numbering plan
	X121                   NumberingPlan // Data numbering plan
	Telex                  NumberingPlan
	ServiceCentreSpecificA NumberingPlan // used to indicate a numbering plan specific to ESME attached to the SMSC
	ServiceCentreSpecificB NumberingPlan // used to indicate a numbering plan specific to ESME attached to the SMSC
	National               NumberingPlan
	Private                NumberingPlan
	ERMES                  NumberingPlan
	Reserved               NumberingPlan // for future extension
}{
	Unknown:                0b0000,
	E164:                   0b0001,
	X121:                   0b0011,
	Telex:                  0b0100,
	ServiceCentreSpecificA: 0b0101,
	ServiceCentreSpecificB: 0b0110,
	National:               0b1000,
	Private:                0b1001,
	ERMES:                  0b1010,
	Reserved:               0b1111,
}

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

// Type returns the type of address (a combination of type-of-number and
// numbering-plan-identification). Currently, only national and
// international E.164 numbers are understood. While ReadFrom() can
// parse alphanumeric numbers, Type() doesn't recognize it.
func (p PhoneNumber) Type() byte {
	typ := PhoneNumberTypes.National
	if strings.HasPrefix(string(p), "+") {
		typ = PhoneNumberTypes.International
	}
	return 0x80 | byte(typ) | byte(NumberingPlans.E164)
}

// ReadFrom constructs an address from the semi-decoded version in the supplied byte slice.
func (p *PhoneNumber) ReadFrom(octets []byte) error {
	if len(octets) < 1 {
		return ErrIncorrectSize
	}

	typ := PhoneNumberType(octets[0] & 0b0111_0000)
	switch typ {
	case PhoneNumberTypes.Alphanumeric:
		addr, err := pdu.Decode7Bit(octets[1:])
		if err != nil {
			return err
		}
		*p = PhoneNumber(addr)
	case PhoneNumberTypes.International:
		addr := pdu.DecodeSemiAddress(octets[1:])
		*p = PhoneNumber("+" + addr)
	case PhoneNumberTypes.National:
		addr := pdu.DecodeSemiAddress(octets[1:])
		*p = PhoneNumber(addr)
	default:
		return fmt.Errorf("%w: Type(0x%x)", ErrUnsupportedTypeOfNumber, typ)
	}
	return nil
}
