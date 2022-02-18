package sms

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
