package sms

import (
	"bytes"
	"io"
)

// Low-level representation of an submit-type SMS message (3GPP TS 23.040).
type smsSubmit struct {
	MessageTypeIndicator    byte
	RejectDuplicates        bool
	ValidityPeriodFormat    byte
	ReplyPath               bool
	UserDataHeaderIndicator bool
	StatusReportRequest     bool

	MessageReference   byte
	DestinationAddress []byte
	ProtocolIdentifier byte
	DataCodingScheme   byte
	ValidityPeriod     []byte
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
		buf.Write(s.ValidityPeriod)
	}
	buf.WriteByte(s.UserDataLength)
	buf.Write(s.UserData)
	return buf.Bytes()
}

func (s *smsSubmit) FromBytes(octets []byte) (n int, err error) { //nolint:funlen
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
		s.ValidityPeriod = make([]byte, 1)
		off, err = io.ReadFull(buf, s.ValidityPeriod)
		n += off
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
	return n, nil
}
