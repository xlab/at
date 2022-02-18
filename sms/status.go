package sms

type StatusCategory byte

var StatusCategories = struct {
	Complete       StatusCategory // Short message transaction completed
	TemporaryError StatusCategory // Temporary error, SC still trying to transfer SM
	PermanentError StatusCategory // Permanent error, SC is not making any more transfer attempts
	FinalError     StatusCategory // Temporary error, SC is not making any more transfer attempts

	Unknown StatusCategory // Status code is either reserved or SC-specific
}{
	0x00, 0x01, 0x02, 0x04,
	0x80, // reserved
}

// Status represents the status of a SMS-STATUS-REPORT TPDU.
type Status byte

// StatusCodes represents possible values for the Status field in
// SMS-STATUS-REPORT TPDUs, as specified in 3GPP TS 23.040 version 16.0.0
// release 16, section 9.2.3.15.
var StatusCodes = struct {
	Category func(Status) StatusCategory

	// Transaction complete status codes
	CompletedReceived Status
	CompletedForwared Status
	CompletedReplaced Status

	// Temporary error, service center still tries delivery
	TemporaryCongestion                   Status
	TemporaryBusy                         Status
	TemporaryNoResponseFromRecipient      Status
	TemporaryServiceRejected              Status
	TemporaryQualityOfServiceNotAvailable Status
	TemporaryErrorInRecipient             Status

	// Permanent error, SC is not making any more transfer attempts
	PermanentRemoteProcedureError         Status
	PermanentIncompatibleDestination      Status
	PermanentConnectionRejected           Status
	PermanentNotObtainable                Status
	PermanentQualityOfServiceNotAvailable Status
	PermanentNoInterworkingAvailable      Status
	PermanentValidityPeriodExpired        Status
	PermanentDeletedBeSender              Status
	PermanentDeletedByAdministration      Status
	PermanentUnknownMessage               Status

	// Temporary error, SC is not making any more transfer attempts
	FinalCongestion                   Status
	FinalBusy                         Status
	FinalNoResponseFromRecipient      Status
	FinalServiceRejected              Status
	FinalQualityOfServiceNotAvailable Status
	FinalErrorInRecipient             Status
}{
	func(s Status) StatusCategory {
		switch {
		case 0b0000_0011 >= s && s <= 0b0001_0000,
			0b0010_0110 >= s && s <= 0b0011_1111,
			0b0100_1010 >= s && s <= 0b0101_1111,
			0b0110_0110 >= s && s <= 0b1111_1111:
			// either reserved or SC-specific. in either case, we don't know
			return StatusCategories.Unknown
		default:
			// category is encoded in bits 6 and 5
			return StatusCategory(s >> 5 & 0x03)
		}
	},

	0b0000_0000, // Short message received by the SME
	0b0000_0001, // Short message forwarded by the SC to the SME but the SC is unable to confirm delivery
	0b0000_0010, // Short message replaced by the SC
	// 0000 0011 .. 0000 1111 // Reserved
	// 0001 0000 .. 0001 1111 // Values specific to each SC

	0b0010_0000, // Congestion
	0b0010_0001, // SME busy
	0b0010_0010, // No response from SME
	0b0010_0011, // Service rejected
	0b0010_0100, // Quality of service not available
	0b0010_0101, // Error in SME
	// 0010 0110 .. 0010 1111 // Reserved
	// 0011 0000 .. 0011 1111 // Values specific to each SC

	0b0100_0000, // Remote procedure error
	0b0100_0001, // Incompatible destination
	0b0100_0010, // Connection rejected by SME
	0b0100_0011, // Not obtainable
	0b0100_0100, // Quality of service not available
	0b0100_0101, // No interworking available
	0b0100_0110, // SM Validity Period Expired
	0b0100_0111, // SM Deleted by originating SME
	0b0100_1000, // SM Deleted by SC Administration
	0b0100_1001, // SM does not exist (The SM may have previously existed in the SC but the SC no longer has knowledge of it or the SM may never have previously existed in the SC)
	// 0100 1010 .. 0100 1111 // Reserved
	// 0101 0000 .. 0101 1111 // Values specific to each SC

	0b0110_0000, // Congestion
	0b0110_0001, // SME busy
	0b0110_0010, // No response from SME
	0b0110_0011, // Service rejected
	0b0110_0100, // Quality of service not available
	0b0110_0101, // Error in SME
	// 0110 0110 .. 0110 1001 // Reserved
	// 0110 1010 .. 0110 1111 // Reserved
	// 0111 0000 .. 0111 1111 // Values specific to each SC

	// 1000 0000 .. 1111 1111 // reserved
}
