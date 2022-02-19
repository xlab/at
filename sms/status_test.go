package sms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusCategorys(t *testing.T) {
	t.Parallel()
	run := func(t *testing.T, name string, cat StatusCategory, s []Status) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for _, status := range s {
				actual := status.Category()
				assert.Equal(t, cat, actual)
			}
		})
	}
	run(t, "complete", StatusCategories.Complete, []Status{
		StatusCodes.CompletedReceived,
		StatusCodes.CompletedForwared,
		StatusCodes.CompletedReplaced,
	})
	run(t, "temporary", StatusCategories.TemporaryError, []Status{
		StatusCodes.TemporaryCongestion,
		StatusCodes.TemporaryBusy,
		StatusCodes.TemporaryNoResponseFromRecipient,
		StatusCodes.TemporaryServiceRejected,
		StatusCodes.TemporaryQualityOfServiceNotAvailable,
		StatusCodes.TemporaryErrorInRecipient,
	})
	run(t, "permanent", StatusCategories.PermanentError, []Status{
		StatusCodes.PermanentRemoteProcedureError,
		StatusCodes.PermanentIncompatibleDestination,
		StatusCodes.PermanentConnectionRejected,
		StatusCodes.PermanentNotObtainable,
		StatusCodes.PermanentQualityOfServiceNotAvailable,
		StatusCodes.PermanentNoInterworkingAvailable,
		StatusCodes.PermanentValidityPeriodExpired,
		StatusCodes.PermanentDeletedBeSender,
		StatusCodes.PermanentDeletedByAdministration,
		StatusCodes.PermanentUnknownMessage,
	})
	run(t, "final", StatusCategories.FinalError, []Status{
		StatusCodes.FinalCongestion,
		StatusCodes.FinalBusy,
		StatusCodes.FinalNoResponseFromRecipient,
		StatusCodes.FinalServiceRejected,
		StatusCodes.FinalQualityOfServiceNotAvailable,
		StatusCodes.FinalErrorInRecipient,
	})
	t.Run("unknown", func(t *testing.T) {
		for _, ranges := range []struct{ begin, end byte }{
			{0b0000_0011, 0b0000_1111}, // complete: Reserved
			{0b0001_0000, 0b0001_1111}, // complete: Values specific to each SC
			{0b0010_0110, 0b0010_1111}, // temporary: Reserved
			{0b0011_0000, 0b0011_1111}, // temporary: Values specific to each SC
			{0b0100_1010, 0b0100_1111}, // permanent: Reserved
			{0b0101_0000, 0b0101_1111}, // permanent: Values specific to each SC
			{0b0110_0110, 0b0110_1001}, // final: Reserved
			{0b0110_1010, 0b0110_1111}, // final: Reserved
			{0b0111_0000, 0b0111_1111}, // final: Values specific to each SC
			{0b1000_0000, 0b1111_1111}, // extension: reserved
		} {
			for i := ranges.begin; i < ranges.end; i++ {
				actual := Status(i).Category()
				assert.Equal(t, StatusCategories.Unknown, actual, "Status(%08b)", i)
			}
		}
	})
}
