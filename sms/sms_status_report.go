package sms

import (
	"bytes"
	"io"
)

// Low-level representation of an status-report-type SMS message (3GPP TS 23.040).
type smsStatusReport struct {
	MessageTypeIndicator     byte
	MoreMessagesToSend       bool
	LoopPrevention           bool
	UserDataHeaderIndicator  bool
	StatusReportQualificator bool

	MessageReference       byte
	DestinationAddress     []byte
	Parameters             byte
	ProtocolIdentifier     byte
	DataCodingScheme       byte
	ServiceCentreTimestamp []byte
	DischargeTimestamp     []byte
	Status                 byte
	UserDataLength         byte
	UserData               []byte
}

func (s *smsStatusReport) Bytes() []byte {
	var buf bytes.Buffer
	header := s.MessageTypeIndicator // 0-1 bits
	if !s.MoreMessagesToSend {
		header |= 0x01 << 2 // 2 bit
	}
	if s.LoopPrevention {
		header |= 0x01 << 3 // 3 bit
	}
	if s.StatusReportQualificator {
		header |= 0x01 << 5 // 5 bit
	}
	if s.UserDataHeaderIndicator {
		header |= 0x01 << 6 // 6 bit
	}
	buf.WriteByte(header)
	buf.WriteByte(s.MessageReference)
	buf.Write(s.DestinationAddress)
	buf.Write(s.ServiceCentreTimestamp)
	buf.Write(s.DischargeTimestamp)
	buf.WriteByte(s.Status)

	var trailer bytes.Buffer
	var indicator byte
	if s.ProtocolIdentifier != 0 {
		indicator |= 0x01 << 0 // 0 bit
		trailer.WriteByte(s.ProtocolIdentifier)
	}
	if s.DataCodingScheme != 0 {
		indicator |= 0x01 << 1 // 1 bit
		trailer.WriteByte(s.DataCodingScheme)
	}
	if s.UserDataHeaderIndicator {
		indicator |= 0x01 << 2 // 2 bit
		trailer.WriteByte(s.UserDataLength)
		trailer.Write(s.UserData)
	}
	buf.WriteByte(indicator)
	if indicator != 0 {
		trailer.WriteTo(&buf)
	}
	return buf.Bytes()
}

func (s *smsStatusReport) FromBytes(octets []byte) (n int, err error) { //nolint:funlen
	buf := bytes.NewReader(octets)
	*s = smsStatusReport{}
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
		s.StatusReportQualificator = true
	}
	s.UserDataHeaderIndicator = header&(0x01<<6) != 0

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
	buf.UnreadByte() // will read length again
	n--
	s.DestinationAddress = make([]byte, blocks(int(daLen), 2)+2)
	off, err := io.ReadFull(buf, s.DestinationAddress)
	n += off
	if err != nil {
		return
	}
	s.ServiceCentreTimestamp = make([]byte, 7)
	off, err = io.ReadFull(buf, s.ServiceCentreTimestamp)
	n += off
	if err != nil {
		return
	}
	s.DischargeTimestamp = make([]byte, 7)
	off, err = io.ReadFull(buf, s.DischargeTimestamp)
	n += off
	if err != nil {
		return
	}
	s.Status, err = buf.ReadByte()
	n++
	if err != nil {
		return
	}
	s.Parameters, err = buf.ReadByte()
	n++
	if err != nil {
		return n - 1, nil
	}
	if s.Parameters&0x01 != 0 {
		s.ProtocolIdentifier, err = buf.ReadByte()
		n++
		if err != nil {
			return
		}
	}
	if s.Parameters&0x02 != 0 {
		s.DataCodingScheme, err = buf.ReadByte()
		n++
		if err != nil {
			return
		}
	}
	if s.Parameters&0x04 != 0 {
		s.UserDataLength, err = buf.ReadByte()
		n++
		if err != nil {
			return
		}
		s.UserData = make([]byte, int(s.UserDataLength))
		off, _ = io.ReadFull(buf, s.UserData)
		s.UserData = s.UserData[:off]
		n += off
	}
	return n, err
}
