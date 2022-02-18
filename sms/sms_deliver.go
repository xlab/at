package sms

import (
	"bytes"
	"io"
)

// Low-level representation of an deliver-type SMS message (3GPP TS 23.040).
type smsDeliver struct {
	MessageTypeIndicator    byte
	MoreMessagesToSend      bool
	LoopPrevention          bool
	ReplyPath               bool
	UserDataHeaderIndicator bool
	StatusReportIndication  bool

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
// The TP-User-Data-Header-Indicator is a 1 bit field within bit 6 of the first octet of an
// SMS-SUBMIT and SMS-DELIVER PDU and has the following values.
// Bit no. 6 0 The TP-UD field contains only the short message 1 The beginning of the TP-UD
// field contains a Header in addition to the short message

// The TP-Reply-Path is a 1-bit field, located within bit no 7 of the first octet of both
// SMS-DELIVER and SMS-SUBMIT, and to be given the following values:
// Bit no 7: 0 TP-Reply-Path parameter is not set in this SMS-SUBMIT/DELIVER
// 1 TP-Reply-Path parameter is set in this SMS-SUBMIT/DELIVER

// TP-OA TP-Originating-Address  2-12 octets
// Each address field of the SM-TL consists of the following sub-fields: An Address-Length
// field of one octet, a Type-of-Address field of one octet, and one Address-Value field
// of variable length
//
// The Address-Length field is an integer representation of the number of useful semi-octets
// within the Address-Value field, i.e. excludes any semi octet containing only fill bits.

func (s *smsDeliver) FromBytes(octets []byte) (n int, err error) { //nolint:funlen
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
	return n, nil
}
