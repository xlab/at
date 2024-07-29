package sms

// Encoding represents the encoding of message's text data.
type Encoding byte

// Encodings represent the possible encodings of message's text data.
var Encodings = struct {
	Gsm7Bit   Encoding
	UCS2      Encoding
	Gsm7Bit_2 Encoding
	Gsm7Bit_3 Encoding
}{
	0x00, 0x08, 0x11, 0x01,
}
