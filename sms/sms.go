// Package sms allows to encode and decode SMS messages into/from PDU format as described in 3GPP TS 23.040.
package sms

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/xlab/at/pdu"
)

// Common errors.
var (
	ErrUnknownEncoding    = errors.New("sms: unsupported encoding")
	ErrUnknownMessageType = errors.New("sms: unsupported message type")
	ErrIncorrectSize      = errors.New("sms: decoded incorrect size of field")
	ErrNonRelative        = errors.New("sms: non-relative validity period support is not implemented yet")
)

// MessageType represents the message's type.
type MessageType byte

// MessageTypes represent the possible message's types (3GPP TS 23.040).
var MessageTypes = struct {
	Deliver       MessageType
	DeliverReport MessageType
	StatusReport  MessageType
	Command       MessageType
	Submit        MessageType
	SubmitReport  MessageType
}{
	0x00, 0x00,
	0x02, 0x02,
	0x01, 0x01,
}

// ValidityPeriodFormat represents the format of message's validity period.
type ValidityPeriodFormat byte

// ValidityPeriodFormats represent the possible formats of message's
// validity period (3GPP TS 23.040).
var ValidityPeriodFormats = struct {
	FieldNotPresent ValidityPeriodFormat
	Relative        ValidityPeriodFormat
	Enhanced        ValidityPeriodFormat
	Absolute        ValidityPeriodFormat
}{
	0x00, 0x02, 0x01, 0x03,
}

// Encoding represents the encoding of message's text data.
type Encoding byte

// Encodings represent the possible encodings of message's text data.
var Encodings = struct {
	Gsm7Bit   Encoding
	UCS2      Encoding
	Gsm7Bit_2 Encoding
}{
	0x00, 0x08, 0x11,
}

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

// USSD represents an USSD query string
type USSD string

// Gsm7Bit encodes USSD query into GSM 7-Bit packed octets.
func (u USSD) Gsm7Bit() []byte {
	return pdu.Encode7Bit(string(u))
}

// ValidityPeriod represents the validity period of message.
type ValidityPeriod time.Duration

// Octet return a one-byte representation of the validity period.
func (v ValidityPeriod) Octet() byte {
	switch d := time.Duration(v); {
	case d/time.Minute < 5:
		return 0x00
	case d/time.Hour < 12:
		return byte(d / (time.Minute * 5))
	case d/time.Hour < 24:
		return byte((d-d/time.Hour*12)/(time.Minute*30) + 143)
	case d/time.Hour < 744:
		days := d / (time.Hour * 24)
		return byte(days + 166)
	default:
		weeks := d / (time.Hour * 24 * 7)
		if weeks > 62 {
			return 0xFF
		}
		return byte(weeks + 192)
	}
}

// ReadFrom reads the validity period form the given byte.
func (v *ValidityPeriod) ReadFrom(oct byte) {
	switch n := time.Duration(oct); {
	case n >= 0 && n <= 143:
		*v = ValidityPeriod(5 * time.Minute * n)
	case n >= 144 && n <= 167:
		*v = ValidityPeriod(12*time.Hour + 30*time.Minute*(n-143))
	case n >= 168 && n <= 196:
		*v = ValidityPeriod(24 * time.Hour * (n - 166))
	case n >= 197 && n <= 255:
		*v = ValidityPeriod(7 * 24 * time.Hour * (n - 192))
	}
}

// Timestamp represents message's timestamp.
type Timestamp time.Time

// GSM 03.**
// TP-Service-Centre-Time-Stamp (TP-SCTS)
//
//  |              | Year | Month | Day | Hour | Minute | Second | Time Zone |
//  |--------------|------|-------|-----|------|--------|--------|-----------|
//  |  Semi-octets |   2  |   2   |  2  |   2  |    2   |    2   |     2     |
//
//  The Time Zone indicates the difference, expressed in quarters of an hour, between the local time and
//  GMT. In the first of the two semi-octets, the first bit (bit 3 of the seventh octet of the TP-Service-CentreTime-Stamp
//  field) represents the algebraic sign of this difference (0 : positive, 1 : negative).

// PDU returns bytes of semi-octet encoded timestamp.
func (t Timestamp) PDU() []byte {
	date := time.Time(t)
	year := date.Year()
	month := date.Month()
	day := date.Day()
	hour := date.Hour()
	minute := date.Minute()
	second := date.Second()

	_, offset := date.Zone()
	negativeOffset := offset < 0
	if negativeOffset {
		offset = -offset
	}
	quarters := offset / int(time.Hour/time.Second) * 4

	_year := pdu.Swap(pdu.Encode((year % 1000)))
	_month := pdu.Swap(pdu.Encode(int(month)))
	_day := pdu.Swap(pdu.Encode(day))
	_hour := pdu.Swap(pdu.Encode(hour))
	_minute := pdu.Swap(pdu.Encode(minute))
	_second := pdu.Swap(pdu.Encode(second))
	_quarters := pdu.Swap(pdu.Encode(quarters))
	if negativeOffset {
		_quarters = _quarters | 0x04
	}

	return []byte{_year, _month, _day, _hour, _minute, _second, _quarters}
}

// ReadFrom reads a semi-encoded timestamp from the given octets.
func (t *Timestamp) ReadFrom(octets []byte) {
	millenium := (time.Now().Year() / 1000) * 1000
	year := pdu.Decode(pdu.Swap(octets[0]))
	month := pdu.Decode(pdu.Swap(octets[1]))
	day := pdu.Decode(pdu.Swap(octets[2]))
	hour := pdu.Decode(pdu.Swap(octets[3]))
	minute := pdu.Decode(pdu.Swap(octets[4]))
	second := pdu.Decode(pdu.Swap(octets[5]))

	negativeOffset := (octets[6] & 0x04) != 0
	quarters := pdu.Decode(pdu.Swap(octets[6] & 0xF7))
	offset := time.Duration(quarters) * 15 * time.Minute

	date := time.Date(millenium+year, time.Month(month), day, hour, minute, second, 0, time.UTC)

	if negativeOffset {
		// was negative, so make UTC
		date = date.Add(offset)
	} else {
		// was positive, so make UTC
		date = date.Add(-offset)
	}
	*t = Timestamp(date.In(time.Local))
}

// Message represents an SMS message, including some advanced fields. This
// is a user-friendly high-level representation that should be used around.
// Complies with 3GPP TS 23.040.
type Message struct {
	Type                 MessageType
	Encoding             Encoding
	VP                   ValidityPeriod
	VPFormat             ValidityPeriodFormat
	ServiceCenterTime    Timestamp
	ServiceCenterAddress PhoneNumber
	Address              PhoneNumber
	Text                 string

	// Advanced
	MessageReference         byte
	ReplyPathExists          bool
	UserDataStartsWithHeader bool
	StatusReportIndication   bool
	StatusReportRequest      bool
	MoreMessagesToSend       bool
	LoopPrevention           bool
	RejectDuplicates         bool
}

func blocks(n, block int) int {
	if n%block == 0 {
		return n / block
	}
	return n/block + 1
}

// PDU serializes the message into octets ready to be transferred.
// Returns the number of TPDU bytes in the produced PDU.
// Complies with 3GPP TS 23.040.
func (s *Message) PDU() (int, []byte, error) {
	var buf bytes.Buffer
	if len(s.ServiceCenterAddress) < 1 {
		buf.WriteByte(0x00) // SMSC info length
	} else {
		_, octets, err := s.ServiceCenterAddress.PDU()
		if err != nil {
			return 0, nil, err
		}
		buf.WriteByte(byte(len(octets)))
		buf.Write(octets)
	}

	switch s.Type {
	case MessageTypes.Deliver:
		var sms smsDeliver
		sms.MessageTypeIndicator = byte(s.Type)
		sms.MoreMessagesToSend = s.MoreMessagesToSend
		sms.LoopPrevention = s.LoopPrevention
		sms.ReplyPath = s.ReplyPathExists
		sms.UserDataHeaderIndicator = s.UserDataStartsWithHeader
		sms.StatusReportIndication = s.StatusReportIndication

		addrLen, addr, err := s.Address.PDU()
		if err != nil {
			return 0, nil, err
		}
		var addrBuf bytes.Buffer
		addrBuf.WriteByte(byte(addrLen))
		addrBuf.Write(addr)
		sms.OriginatingAddress = addrBuf.Bytes()

		sms.ProtocolIdentifier = 0x00 // Short Message Type 0
		sms.DataCodingScheme = byte(s.Encoding)
		sms.ServiceCentreTimestamp = s.ServiceCenterTime.PDU()

		var userData []byte
		switch s.Encoding {
		case Encodings.Gsm7Bit, Encodings.Gsm7Bit_2:
			userData = pdu.Encode7Bit(s.Text)
			sms.UserDataLength = byte(len(s.Text))
		case Encodings.UCS2:
			userData = pdu.EncodeUcs2(s.Text)
			sms.UserDataLength = byte(len(userData))
		default:
			return 0, nil, ErrUnknownEncoding
		}

		sms.UserData = userData
		n, err := buf.Write(sms.Bytes())
		if err != nil {
			return 0, nil, err
		}
		return n, buf.Bytes(), nil
	case MessageTypes.Submit:
		var sms smsSubmit
		sms.MessageTypeIndicator = byte(s.Type)
		sms.RejectDuplicates = s.RejectDuplicates
		sms.ValidityPeriodFormat = byte(s.VPFormat)
		sms.ReplyPath = s.ReplyPathExists
		sms.UserDataHeaderIndicator = s.UserDataStartsWithHeader
		sms.StatusReportRequest = s.StatusReportRequest
		sms.MessageReference = s.MessageReference

		addrLen, addr, err := s.Address.PDU()
		if err != nil {
			return 0, nil, err
		}
		var addrBuf bytes.Buffer
		addrBuf.WriteByte(byte(addrLen))
		addrBuf.Write(addr)
		sms.DestinationAddress = addrBuf.Bytes()

		sms.ProtocolIdentifier = 0x00 // Short Message Type 0
		sms.DataCodingScheme = byte(s.Encoding)

		switch s.VPFormat {
		case ValidityPeriodFormats.Relative:
			sms.ValidityPeriod = byte(s.VP.Octet())
		case ValidityPeriodFormats.Absolute, ValidityPeriodFormats.Enhanced:
			return 0, nil, ErrNonRelative
		}

		var userData []byte
		switch s.Encoding {
		case Encodings.Gsm7Bit, Encodings.Gsm7Bit_2:
			userData = pdu.Encode7Bit(s.Text)
			sms.UserDataLength = byte(len(s.Text))
		case Encodings.UCS2:
			userData = pdu.EncodeUcs2(s.Text)
			sms.UserDataLength = byte(len(userData))
		default:
			return 0, nil, ErrUnknownEncoding
		}

		sms.UserData = userData
		n, err := buf.Write(sms.Bytes())
		if err != nil {
			return 0, nil, err
		}
		return n, buf.Bytes(), nil
	default:
		return 0, nil, ErrUnknownMessageType
	}
}

// ReadFrom constructs a message from the supplied PDU octets. Returns the number of bytes read.
// Complies with 3GPP TS 23.040.
func (s *Message) ReadFrom(octets []byte) (n int, err error) {
	*s = Message{}
	buf := bytes.NewReader(octets)
	scLen, err := buf.ReadByte()
	n++
	if err != nil {
		return
	}
	if scLen > 16 {
		return 0, ErrIncorrectSize
	}
	addr := make([]byte, scLen)
	off, err := io.ReadFull(buf, addr)
	n += off
	if err != nil {
		return
	}
	s.ServiceCenterAddress.ReadFrom(addr)
	msgType, err := buf.ReadByte()
	n++
	if err != nil {
		return
	}
	n--
	buf.UnreadByte()
	s.Type = MessageType(msgType & 0x03)

	switch s.Type {
	case MessageTypes.Deliver:
		var sms smsDeliver
		off, err2 := sms.FromBytes(octets[1+scLen:])
		n += off
		if err2 != nil {
			return n, err2
		}
		s.MoreMessagesToSend = sms.MoreMessagesToSend
		s.LoopPrevention = sms.LoopPrevention
		s.ReplyPathExists = sms.ReplyPath
		s.UserDataStartsWithHeader = sms.UserDataHeaderIndicator
		s.StatusReportIndication = sms.StatusReportIndication
		s.Address.ReadFrom(sms.OriginatingAddress[1:])
		s.Encoding = Encoding(sms.DataCodingScheme)
		s.ServiceCenterTime.ReadFrom(sms.ServiceCentreTimestamp)
		switch s.Encoding {
		case Encodings.Gsm7Bit, Encodings.Gsm7Bit_2:
			s.Text, err = pdu.Decode7Bit(sms.UserData)
			if err != nil {
				return
			}
			s.Text = cutStr(s.Text, int(sms.UserDataLength))
		case Encodings.UCS2:
			s.Text, err = pdu.DecodeUcs2(sms.UserData, s.UserDataStartsWithHeader)
			if err != nil {
				return
			}
		default:
			return 0, ErrUnknownEncoding
		}
	case MessageTypes.Submit:
		var sms smsSubmit
		off, err2 := sms.FromBytes(octets[1+scLen:])
		n += off
		if err2 != nil {
			return n, err2
		}
		s.RejectDuplicates = sms.RejectDuplicates

		switch s.VPFormat {
		case ValidityPeriodFormats.Absolute, ValidityPeriodFormats.Enhanced:
			return n, ErrNonRelative
		default:
			s.VPFormat = ValidityPeriodFormat(sms.ValidityPeriodFormat)
		}

		s.ReplyPathExists = sms.ReplyPath
		s.UserDataStartsWithHeader = sms.UserDataHeaderIndicator
		s.StatusReportRequest = sms.StatusReportRequest
		s.Address.ReadFrom(sms.DestinationAddress[1:])
		s.Encoding = Encoding(sms.DataCodingScheme)

		if s.VPFormat != ValidityPeriodFormats.FieldNotPresent {
			s.VP.ReadFrom(sms.ValidityPeriod)
		}

		switch s.Encoding {
		case Encodings.Gsm7Bit, Encodings.Gsm7Bit_2:
			s.Text, err = pdu.Decode7Bit(sms.UserData)
			if err != nil {
				return
			}
			s.Text = cutStr(s.Text, int(sms.UserDataLength))
		case Encodings.UCS2:
			s.Text, err = pdu.DecodeUcs2(sms.UserData, s.UserDataStartsWithHeader)
			if err != nil {
				return
			}
		default:
			return 0, ErrUnknownEncoding
		}
	default:
		return n, ErrUnknownMessageType
	}

	return
}

// Low-level representation of an deliver-type SMS message (3GPP TS 23.040).
type smsDeliver struct {
	MessageTypeIndicator    byte
	MoreMessagesToSend      bool
	LoopPrevention          bool
	ReplyPath               bool
	UserDataHeaderIndicator bool
	StatusReportIndication  bool
	// =========================
	OriginatingAddress     []byte
	ProtocolIdentifier     byte
	DataCodingScheme       byte
	ServiceCentreTimestamp []byte
	UserDataLength         byte
	UserData               []byte
}

func (s *smsDeliver) Bytes() []byte {
	var buf bytes.Buffer
	header := s.MessageTypeIndicator // 0-1 bits
	if !s.MoreMessagesToSend {
		header |= 0x01 << 2 // 2 bit
	}
	if s.LoopPrevention {
		header |= 0x01 << 3 // 3 bit
	}
	if s.StatusReportIndication {
		header |= 0x01 << 4 // 4 bit
	}
	if s.UserDataHeaderIndicator {
		header |= 0x01 << 5 // 5 bit
	}
	if s.ReplyPath {
		header |= 0x01 << 6 // 6 bit
	}
	buf.WriteByte(header)
	buf.Write(s.OriginatingAddress)
	buf.WriteByte(s.ProtocolIdentifier)
	buf.WriteByte(s.DataCodingScheme)
	buf.Write(s.ServiceCentreTimestamp)
	buf.WriteByte(s.UserDataLength)
	buf.Write(s.UserData)
	return buf.Bytes()
}

// GSM 03.**
//The TP-User-Data-Header-Indicator is a 1 bit field within bit 6 of the first octet of an SMS-SUBMIT and
//SMS-DELIVER PDU and has the following values.
//Bit no. 6 0 The TP-UD field contains only the short message
//1 The beginning of the TP-UD field contains a Header in addition to the
//short message

//The TP-Reply-Path is a 1-bit field, located within bit no 7 of the first octet of both SMS-DELIVER and
//SMS-SUBMIT, and to be given the following values:
//Bit no 7: 0 TP-Reply-Path parameter is not set in this SMS-SUBMIT/DELIVER
//1 TP-Reply-Path parameter is set in this SMS-SUBMIT/DELIVER

// TP-OA TP-Originating-Address  2-12 octets
// Each address field of the SM-TL consists of the following sub-fields: An Address-Length field of one octet,
// a Type-of-Address field of one octet, and one Address-Value field of variable length
//
// The Address-Length field is an integer representation of the number of useful semi-octets within the
// Address-Value field, i.e. excludes any semi octet containing only fill bits.

func (s *smsDeliver) FromBytes(octets []byte) (n int, err error) {
	buf := bytes.NewReader(octets)
	*s = smsDeliver{}
	header, err := buf.ReadByte()
	n++
	if err != nil {
		return
	}
	s.MessageTypeIndicator = header & 0x03
	if header>>2&0x01 == 0x00 {
		s.MoreMessagesToSend = true
	}
	if header>>3&0x01 == 0x01 {
		s.LoopPrevention = true
	}
	if header>>4&0x01 == 0x01 {
		s.StatusReportIndication = true
	}

	s.UserDataHeaderIndicator = header&(0x01<<6) != 0
	s.ReplyPath = header&(0x01<<7) != 0

	oaLen, err := buf.ReadByte()
	n++
	if err != nil {
		return
	}
	buf.UnreadByte() // will read length again
	n--
	s.OriginatingAddress = make([]byte, blocks(int(oaLen), 2)+2)
	off, err := io.ReadFull(buf, s.OriginatingAddress)
	n += off
	if err != nil {
		return
	}
	s.ProtocolIdentifier, err = buf.ReadByte()
	n++
	if err != nil {
		return
	}
	s.DataCodingScheme, err = buf.ReadByte()
	n++
	if err != nil {
		return
	}
	s.ServiceCentreTimestamp = make([]byte, 7)
	off, err = io.ReadFull(buf, s.ServiceCentreTimestamp)
	n += off
	if err != nil {
		return
	}
	s.UserDataLength, err = buf.ReadByte()
	n++
	if err != nil {
		return
	}
	s.UserData = make([]byte, int(s.UserDataLength))
	off, _ = io.ReadFull(buf, s.UserData)
	s.UserData = s.UserData[:off]
	n += off
	return
}

// Low-level representation of an submit-type SMS message (3GPP TS 23.040).
type smsSubmit struct {
	MessageTypeIndicator    byte
	RejectDuplicates        bool
	ValidityPeriodFormat    byte
	ReplyPath               bool
	UserDataHeaderIndicator bool
	StatusReportRequest     bool
	// =========================
	MessageReference   byte
	DestinationAddress []byte
	ProtocolIdentifier byte
	DataCodingScheme   byte
	ValidityPeriod     byte
	UserDataLength     byte
	UserData           []byte
}

func (s *smsSubmit) Bytes() []byte {
	var buf bytes.Buffer
	header := s.MessageTypeIndicator // 0-1 bits
	if s.RejectDuplicates {
		header |= 0x01 << 2 // 2 bit
	}
	header |= s.ValidityPeriodFormat << 3 // 3-4 bits
	if s.StatusReportRequest {
		header |= 0x01 << 5 // 5 bit
	}
	if s.UserDataHeaderIndicator {
		header |= 0x01 << 6 // 6 bit
	}
	if s.ReplyPath {
		header |= 0x01 << 7 // 7 bit
	}
	buf.WriteByte(header)
	buf.WriteByte(s.MessageReference)
	buf.Write(s.DestinationAddress)
	buf.WriteByte(s.ProtocolIdentifier)
	buf.WriteByte(s.DataCodingScheme)
	if ValidityPeriodFormat(s.ValidityPeriodFormat) != ValidityPeriodFormats.FieldNotPresent {
		buf.WriteByte(s.ValidityPeriod)
	}
	buf.WriteByte(s.UserDataLength)
	buf.Write(s.UserData)
	return buf.Bytes()
}

func (s *smsSubmit) FromBytes(octets []byte) (n int, err error) {
	*s = smsSubmit{}
	buf := bytes.NewReader(octets)
	header, err := buf.ReadByte()
	n++
	if err != nil {
		return
	}
	s.MessageTypeIndicator = header & 0x03
	if header&(0x01<<2) > 0 {
		s.RejectDuplicates = true
	}
	s.ValidityPeriodFormat = header >> 3 & 0x03
	if header&(0x01<<5) > 0 {
		s.StatusReportRequest = true
	}
	if header&(0x01<<6) > 0 {
		s.UserDataHeaderIndicator = true
	}
	if header&(0x01<<7) > 0 {
		s.ReplyPath = true
	}
	s.MessageReference, err = buf.ReadByte()
	n++
	if err != nil {
		return
	}
	daLen, err := buf.ReadByte()
	n++
	if err != nil {
		return
	}
	if daLen > 16 {
		return n, ErrIncorrectSize
	}
	buf.UnreadByte() // read length again
	n--
	s.DestinationAddress = make([]byte, blocks(int(daLen), 2)+2)
	off, err := io.ReadFull(buf, s.DestinationAddress)
	n += off
	if err != nil {
		return
	}
	s.ProtocolIdentifier, err = buf.ReadByte()
	n++
	if err != nil {
		return
	}
	s.DataCodingScheme, err = buf.ReadByte()
	n++
	if err != nil {
		return
	}
	if ValidityPeriodFormat(s.ValidityPeriodFormat) != ValidityPeriodFormats.FieldNotPresent {
		s.ValidityPeriod, err = buf.ReadByte()
		n++
		if err != nil {
			return
		}
	}
	s.UserDataLength, err = buf.ReadByte()
	n++
	if err != nil {
		return
	}
	s.UserData = make([]byte, int(s.UserDataLength))
	off, _ = io.ReadFull(buf, s.UserData)
	s.UserData = s.UserData[:off]
	n += off
	return
}

func cutStr(str string, n int) string {
	runes := []rune(str)
	if n < len(str) {
		return string(runes[0:n])
	}
	return str
}
