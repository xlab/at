package sms

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xlab/at/util"
)

var (
	pduDeliverUCS2 = "07919761989901F0040B919762995696F000084160621263036178042D0442" +
		"043E0442002004300431043E043D0435043D0442002004370432043E043D0438043B0020043" +
		"20430043C0020003200200440043004370430002E0020041F043E0441043B04350434043D04" +
		"3804390020002D002000200032003600200438044E043D044F00200432002000320031003A0" +
		"0330035"
	pduSubmitUCS2 = "07919761989901F011000B919762995696F00008AA78042D0442043E04420020" +
		"04300431043E043D0435043D0442002004370432043E043D0438043B002004320430043C0020" +
		"003200200440043004370430002E0020041F043E0441043B04350434043D043804390020002D" +
		"002000200032003600200438044E043D044F00200432002000320031003A00330035"

	pduDeliverGsm7   = "07919762020033F1040B919762995696F0000041606291401561066379180E8200"
	pduSubmitGsm7    = "07919762020033F111000B919762995696F00000AA066379180E8200"
	pduSubmitGsm7_EnhancedTpVp = "05915155010009010891515511110000420300000000001e547" +
		"47a0e9a36a72074780e9a81e6e5f1db4d9e83e86f103b6d2f03"
	pduDeliverGsm7_2 = "0791551010010201040D91551699296568F80011719022124215293DD4B71C5E26BF" +
		"41D3E6145476D3E5E573BD0C82BF40B59A2D96CBE564351BCE8603A164319D8CA6ABD540E432482673C172AED82DE502"

	pduStatusReport = "079194710600400706360d91947106000000f122206151457440222061514584400000"
)

var (
	smsDeliverUCS2 = Message{
		Text:                 "Этот абонент звонил вам 2 раза. Последний -  26 июня в 21:35",
		Encoding:             Encodings.UCS2,
		Type:                 MessageTypes.Deliver,
		Address:              "+79269965690",
		ServiceCenterAddress: "+79168999100",
		ServiceCenterTime:    parseTimestamp("2014-06-26T21:36:30+04:00"),
	}
	smsDeliverGsm7 = Message{
		Text:                 "crap Δ",
		Encoding:             Encodings.Gsm7Bit,
		Type:                 MessageTypes.Deliver,
		Address:              "+79269965690",
		ServiceCenterAddress: "+79262000331",
		ServiceCenterTime:    parseTimestamp("2014-06-26T19:04:51+04:00"),
	}
	smsDeliverGsm7_2 = Message{
		Text:                 "Torpedo SMS entregue p/ 5561999256868 (21:24:55 de 22.09.17).",
		Encoding:             Encodings.Gsm7Bit_2,
		Type:                 MessageTypes.Deliver,
		Address:              "+5561999256868",
		ServiceCenterAddress: "+550101102010",
		ServiceCenterTime:    parseTimestamp("2017-09-22T21:24:51-03:00"),
	}
	smsSubmitUCS2 = Message{
		Text:                 "Этот абонент звонил вам 2 раза. Последний -  26 июня в 21:35",
		Encoding:             Encodings.UCS2,
		Type:                 MessageTypes.Submit,
		Address:              "+79269965690",
		ServiceCenterAddress: "+79168999100",
		VP:                   ValidityPeriod(time.Hour * 24 * 4),
		VPFormat:             ValidityPeriodFormats.Relative,
	}
	smsSubmitGsm7 = Message{
		Text:                 "crap Δ",
		Encoding:             Encodings.Gsm7Bit,
		Type:                 MessageTypes.Submit,
		Address:              "+79269965690",
		ServiceCenterAddress: "+79262000331",
		VP:                   ValidityPeriod(time.Hour * 24 * 4),
		VPFormat:             ValidityPeriodFormats.Relative,
	}
	smsSubmitGsm7_EnhancedTpVp = Message{
		Text:                 "This SMS has 3 seconds to live",
		Encoding:             Encodings.Gsm7Bit,
		Type:                 MessageTypes.Submit,
		Address:              "+15551111",
		ServiceCenterAddress: "+15551000",
	}
	smsReport = Message{
		Type:                 MessageTypes.StatusReport,
		Status:               0x00, // received
		MessageReference:     54,
		Address:              "+4917600000001",
		ServiceCenterAddress: "+491760000470",
		ServiceCenterTime:    parseTimestamp("2022-02-16T15:54:47+01:00"),
		DischargeTime:        parseTimestamp("2022-02-16T15:54:48+01:00"),
	}
)

// parseTimestamp, a test helper, parses an RFC3339-formatted date into
// a Timestamp. If the input is malformed, parseTimestamp panics.
func parseTimestamp(timetamp string) Timestamp {
	date, err := time.Parse(time.RFC3339, timetamp)
	if err != nil {
		panic(err)
	}
	return Timestamp(date)
}

func TestSmsDeliverReadFromUCS2(t *testing.T) {
	t.Parallel()

	var msg Message
	data, err := util.Bytes(pduDeliverUCS2)
	require.NoError(t, err)
	n, err := msg.ReadFrom(data)
	require.NoError(t, err)
	assert.Equal(t, n, len(data))
	assert.Equal(t, smsDeliverUCS2, msg)
}

func TestSmsDeliverReadFromGsm7(t *testing.T) {
	t.Parallel()

	var msg Message
	data, err := util.Bytes(pduDeliverGsm7)
	require.NoError(t, err)
	n, err := msg.ReadFrom(data)
	require.NoError(t, err)
	assert.Equal(t, n, len(data))
	assert.Equal(t, smsDeliverGsm7, msg)
}

func TestSmsDeliverReadFromGsm7_2(t *testing.T) {
	t.Parallel()

	var msg Message
	data, err := util.Bytes(pduDeliverGsm7_2)
	require.NoError(t, err)
	n, err := msg.ReadFrom(data)
	require.NoError(t, err)
	assert.Equal(t, n, len(data))
	assert.Equal(t, smsDeliverGsm7_2, msg)
}

func TestSmsDeliverPduUCS2(t *testing.T) {
	t.Parallel()

	n, octets, err := smsDeliverUCS2.PDU()
	require.NoError(t, err)
	assert.Equal(t, len(pduDeliverUCS2)/2-8, n)
	data, err := util.Bytes(pduDeliverUCS2)
	require.NoError(t, err)
	assert.Equal(t, data, octets)
}

func TestSmsDeliverPduGsm7(t *testing.T) {
	t.Parallel()

	n, octets, err := smsDeliverGsm7.PDU()
	require.NoError(t, err)
	assert.Equal(t, len(pduDeliverGsm7)/2-8, n)
	data, err := util.Bytes(pduDeliverGsm7)
	t.Logf("%02x\n", string(data))
	t.Logf("%02x\n", string(octets))
	require.NoError(t, err)
	assert.Equal(t, data, octets)
}

func TestSmsSubmitReadFromUCS2(t *testing.T) {
	t.Parallel()

	var msg Message
	data, err := util.Bytes(pduSubmitUCS2)
	require.NoError(t, err)
	n, err := msg.ReadFrom(data)
	require.NoError(t, err)
	assert.Equal(t, n, len(data))
	assert.Equal(t, smsSubmitUCS2, msg)
}

func TestSmsSubmitReadFromGsm7(t *testing.T) {
	t.Parallel()

	var msg Message
	data, err := util.Bytes(pduSubmitGsm7)
	require.NoError(t, err)
	n, err := msg.ReadFrom(data)
	require.NoError(t, err)
	assert.Equal(t, n, len(data))
	assert.Equal(t, smsSubmitGsm7, msg)
}

func TestSmsSubmitReadFromGsm7_EnhancedTpVp(t *testing.T) {
	t.Parallel()

	var msg Message
	data, err := util.Bytes(pduSubmitGsm7_EnhancedTpVp)
	require.NoError(t, err)
	_, err = msg.ReadFrom(data)
	assert.Equal(t, err, ErrNonRelative)
}

func TestSmsSubmitPduUCS2(t *testing.T) {
	t.Parallel()

	n, octets, err := smsSubmitUCS2.PDU()
	require.NoError(t, err)
	assert.Equal(t, len(pduSubmitUCS2)/2-8, n)
	data, err := util.Bytes(pduSubmitUCS2)
	require.NoError(t, err)
	assert.Equal(t, data, octets)
}

func TestSmsSubmitPduGsm7(t *testing.T) {
	t.Parallel()

	n, octets, err := smsSubmitGsm7.PDU()
	require.NoError(t, err)
	assert.Equal(t, len(pduSubmitGsm7)/2-8, n)
	data, err := util.Bytes(pduSubmitGsm7)
	require.NoError(t, err)
	assert.Equal(t, data, octets)
}

func TestSmsStatusReport(t *testing.T) {
	t.Parallel()

	m := Message{}
	_, err := m.ReadFrom(util.MustBytes(pduStatusReport))
	require.NoError(t, err)

	n, octets, err := smsReport.PDU()
	require.NoError(t, err)
	assert.Equal(t, len(pduStatusReport)/2-8, n)
	data, err := util.Bytes(pduStatusReport)
	require.NoError(t, err)
	assert.Equal(t, data, octets)
}
