// Package sms allows to encode and decode SMS messages into/from PDU format as described in 3GPP TS 23.040.
package sms

import (
	"bytes"
	"errors"
	"io"

	"github.com/xlab/at/pdu"
)

// Common errors.
var (
	ErrUnknownEncoding               = errors.New("sms: unsupported encoding")
	ErrUnknownMessageType            = errors.New("sms: unsupported message type")
	ErrIncorrectSize                 = errors.New("sms: decoded incorrect size of field")
	ErrNonRelative                   = errors.New("sms: non-relative validity period support is not implemented yet")
	ErrIncorrectUserDataHeaderLength = errors.New("sms: incorrect user data header length ")
	ErrUnsupportedTypeOfNumber       = errors.New("sms: unsupported type-of-number")
)

// Message represents an SMS message, including some advanced fields. This
// is a user-friendly high-level representation that should be used around.
// Complies with 3GPP TS 23.040.
type Message struct {
	Type                 MessageType
	Encoding             Encoding
	VP                   ValidityPeriod
	VPFormat             ValidityPeriodFormat
	ServiceCenterTime    Timestamp
	DischargeTime        Timestamp
	ServiceCenterAddress PhoneNumber
	Address              PhoneNumber
	Text                 string
	UserDataHeader       UserDataHeader

	// Advanced
	MessageReference         byte
	Status                   Status
	ReplyPathExists          bool
	UserDataStartsWithHeader bool
	StatusReportIndication   bool
	StatusReportRequest      bool
	StatusReportQualificator bool
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

func cutStr(str string, n int) string {
	runes := []rune(str)
	if n < len(str) {
		return string(runes[0:n])
	}
	return str
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

	var n int
	var err error

	switch s.Type {
	case MessageTypes.Deliver:
		n, err = s.encodeDeliver(&buf)
	case MessageTypes.Submit:
		n, err = s.encodeSubmit(&buf)
	case MessageTypes.StatusReport:
		n, err = s.encodeStatusReport(&buf)
	default:
		err = ErrUnknownMessageType
	}

	if err != nil {
		return 0, nil, err
	}
	return n, buf.Bytes(), nil
}

func (s *Message) encodeDeliver(buf *bytes.Buffer) (n int, err error) {
	var sms smsDeliver
	sms.MessageTypeIndicator = byte(s.Type)
	sms.MoreMessagesToSend = s.MoreMessagesToSend
	sms.LoopPrevention = s.LoopPrevention
	sms.ReplyPath = s.ReplyPathExists
	sms.UserDataHeaderIndicator = s.UserDataStartsWithHeader
	sms.StatusReportIndication = s.StatusReportIndication

	addrLen, addr, err := s.Address.PDU()
	if err != nil {
		return 0, err
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
		return 0, ErrUnknownEncoding
	}

	sms.UserData = userData
	return buf.Write(sms.Bytes())
}

func (s *Message) encodeSubmit(buf *bytes.Buffer) (n int, err error) {
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
		return 0, err
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
		return 0, ErrNonRelative
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
		return 0, ErrUnknownEncoding
	}

	sms.UserData = userData
	return buf.Write(sms.Bytes())
}

func (s *Message) encodeStatusReport(buf *bytes.Buffer) (n int, err error) {
	var sms smsStatusReport
	sms.MessageTypeIndicator = byte(s.Type)
	sms.UserDataHeaderIndicator = s.UserDataStartsWithHeader
	sms.MoreMessagesToSend = s.MoreMessagesToSend
	sms.LoopPrevention = s.LoopPrevention
	sms.StatusReportQualificator = s.StatusReportQualificator
	sms.MessageReference = s.MessageReference

	addrLen, addr, err := s.Address.PDU()
	if err != nil {
		return 0, err
	}
	var addrBuf bytes.Buffer
	addrBuf.WriteByte(byte(addrLen))
	addrBuf.Write(addr)
	sms.DestinationAddress = addrBuf.Bytes()

	sms.ServiceCentreTimestamp = s.ServiceCenterTime.PDU()
	sms.DischargeTimestamp = s.DischargeTime.PDU()
	sms.Status = byte(s.Status)

	var userData []byte
	switch s.Encoding {
	case Encodings.Gsm7Bit, Encodings.Gsm7Bit_2:
		userData = pdu.Encode7Bit(s.Text)
		sms.UserDataLength = byte(len(s.Text))
	case Encodings.UCS2:
		userData = pdu.EncodeUcs2(s.Text)
		sms.UserDataLength = byte(len(userData))
	default:
		return 0, ErrUnknownEncoding
	}

	sms.UserData = userData
	return buf.Write(sms.Bytes())
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

	var decBytes int
	octets = octets[1+scLen:]

	switch s.Type {
	case MessageTypes.Deliver:
		decBytes, err = s.decodeDeliver(octets)
	case MessageTypes.Submit:
		decBytes, err = s.decodeSubmit(octets)
	case MessageTypes.StatusReport:
		decBytes, err = s.decodeStatusReport(octets)
	default:
		return n, ErrUnknownMessageType
	}

	n += decBytes
	return n, err
}

func (s *Message) decodeDeliver(data []byte) (n int, err error) {
	var sms smsDeliver
	n, err = sms.FromBytes(data)
	if err != nil {
		return
	}
	s.MoreMessagesToSend = sms.MoreMessagesToSend
	s.LoopPrevention = sms.LoopPrevention
	s.ReplyPathExists = sms.ReplyPath
	s.UserDataStartsWithHeader = sms.UserDataHeaderIndicator
	if sms.UserDataHeaderIndicator {
		err = s.UserDataHeader.ReadFrom(sms.UserData)
		if err != nil {
			return
		}
	}
	s.StatusReportIndication = sms.StatusReportIndication
	s.Address.ReadFrom(sms.OriginatingAddress[1:])
	s.Encoding = Encoding(sms.DataCodingScheme)
	s.ServiceCenterTime.ReadFrom(sms.ServiceCentreTimestamp)
	err = s.decodeUserData(sms.UserData, sms.UserDataLength)
	return n, err
}

func (s *Message) decodeSubmit(data []byte) (n int, err error) {
	var sms smsSubmit
	n, err = sms.FromBytes(data)
	if err != nil {
		return
	}
	s.RejectDuplicates = sms.RejectDuplicates

	switch s.VPFormat {
	case ValidityPeriodFormats.Absolute, ValidityPeriodFormats.Enhanced:
		return n, ErrNonRelative
	default:
		s.VPFormat = ValidityPeriodFormat(sms.ValidityPeriodFormat)
	}

	s.MessageReference = sms.MessageReference
	s.ReplyPathExists = sms.ReplyPath
	s.UserDataStartsWithHeader = sms.UserDataHeaderIndicator
	s.StatusReportRequest = sms.StatusReportRequest
	s.Address.ReadFrom(sms.DestinationAddress[1:])
	s.Encoding = Encoding(sms.DataCodingScheme)

	if s.VPFormat != ValidityPeriodFormats.FieldNotPresent {
		s.VP.ReadFrom(sms.ValidityPeriod)
	}
	err = s.decodeUserData(sms.UserData, sms.UserDataLength)
	return n, err
}

func (s *Message) decodeStatusReport(data []byte) (n int, err error) {
	var sms smsStatusReport
	n, err = sms.FromBytes(data)
	if err != nil {
		return
	}
	s.MessageReference = sms.MessageReference
	s.MoreMessagesToSend = sms.MoreMessagesToSend
	s.LoopPrevention = sms.LoopPrevention
	s.UserDataStartsWithHeader = sms.UserDataHeaderIndicator
	if sms.UserDataHeaderIndicator {
		err = s.UserDataHeader.ReadFrom(sms.UserData)
		if err != nil {
			return
		}
	}
	s.StatusReportQualificator = sms.StatusReportQualificator
	s.Status = Status(sms.Status)
	s.Address.ReadFrom(sms.DestinationAddress[1:])
	s.Encoding = Encoding(sms.DataCodingScheme)
	s.ServiceCenterTime.ReadFrom(sms.ServiceCentreTimestamp)
	s.DischargeTime.ReadFrom(sms.DischargeTimestamp)
	err = s.decodeUserData(sms.UserData, sms.UserDataLength)
	return n, err
}

func (s *Message) decodeUserData(data []byte, dataLen byte) (err error) {
	switch s.Encoding {
	case Encodings.Gsm7Bit, Encodings.Gsm7Bit_2:
		if s.Text, err = pdu.Decode7Bit(data); err != nil {
			return
		}
		s.Text = cutStr(s.Text, int(dataLen))
	case Encodings.UCS2:
		s.Text, err = pdu.DecodeUcs2(data, s.UserDataStartsWithHeader)
	default:
		return ErrUnknownEncoding
	}
	return err
}
