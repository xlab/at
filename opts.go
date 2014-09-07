package at

import "strings"

// Opt represents a numerical option.
type Opt struct {
	// ID is usually used to detect an option from a number in string.
	ID int
	// Description contains a human-readable description of an option.
	Description string
}

// StringOpt represents a string option.
type StringOpt struct {
	// ID is usually used to detect an option from a substring in string.
	ID string
	// Description contains a human-readable description of an option.
	Description string
}

// UnknownOpt represents an option that was parsed incorrectly or was not parsed at all.
var UnknownOpt = Opt{ID: -1, Description: "-"}

// UnknownStringOpt represents a string option that was parsed incorrectly or was not parsed at all.
var UnknownStringOpt = StringOpt{ID: "nil", Description: "Unknown"}

// KillCmd is an artifical AT command that may be successfully sent to device in order
// to emulate the responce from it. In other words, if a connection with device stalled and
// no bytes could be read, then this command is used to read something and then close the connection.
const KillCmd = "AT_KILL"

// NoopCmd is like a ping command that signals that the device is responsive.
const NoopCmd = "AT"

type optMap map[int]Opt
type stringOpts []StringOpt

func (o optMap) Resolve(id int) Opt {
	if opt, ok := o[id]; ok {
		return opt
	}
	return UnknownOpt
}

func (s stringOpts) Resolve(str string) StringOpt {
	for _, v := range s {
		if strings.HasPrefix(str, v.ID) {
			return v
		}
	}
	return UnknownStringOpt
}

// DeviceState represents the device state including cellular options,
// signal quality, current operator name, service status.
type DeviceState struct {
	ServiceState   Opt
	ServiceDomain  Opt
	RoamingState   Opt
	SystemMode     Opt
	SystemSubmode  Opt
	SimState       Opt
	ModelName      string
	OperatorName   string
	IMEI           string
	SignalStrength int
}

// NewDeviceState returns a clean state with unknown options.
func NewDeviceState() *DeviceState {
	return &DeviceState{
		ServiceState:  UnknownOpt,
		ServiceDomain: UnknownOpt,
		RoamingState:  UnknownOpt,
		SystemMode:    UnknownOpt,
		SystemSubmode: UnknownOpt,
		SimState:      UnknownOpt,
	}
}

var sim = optMap{
	0:   Opt{0, "Invalid USIM card or pin code locked"},
	1:   Opt{1, "Valid USIM card"},
	2:   Opt{2, "USIM is invalid for cellular service"},
	3:   Opt{3, "USIM is invalid for packet service"},
	4:   Opt{4, "USIM is not valid for cellular nor packet services"},
	255: Opt{255, "USIM card is not exist"},
}

// SimStates represent the possible data card states.
var SimStates = struct {
	Resolve func(int) Opt

	Invalid     Opt
	Valid       Opt
	InvalidCS   Opt
	InvalidPS   Opt
	InvalidCSPS Opt
	NoCard      Opt
}{
	func(id int) Opt { return sim.Resolve(id) },

	sim[0], sim[1], sim[2], sim[3], sim[4], sim[255],
}

var service = optMap{
	0: Opt{0, "No service"},
	1: Opt{1, "Restricted service"},
	2: Opt{2, "Valid service"},
	3: Opt{3, "Restricted regional service"},
	4: Opt{4, "Power-saving and deep sleep state"},
}

// ServiceStates represent the possible service states.
var ServiceStates = struct {
	Resolve func(int) Opt

	None               Opt
	Restricted         Opt
	Valid              Opt
	RestrictedRegional Opt
	PowerSaving        Opt
}{
	func(id int) Opt { return service.Resolve(id) },

	service[0], service[1], service[2], service[3], service[4],
}

var domain = optMap{
	0: Opt{0, "No service"},
	1: Opt{1, "Cellular service only"},
	2: Opt{2, "Packet service only"},
	3: Opt{3, "Packet and Cellular services"},
	4: Opt{4, "Searching"},
}

// ServiceDomains represent the possible service domains.
var ServiceDomains = struct {
	Resolve func(int) Opt

	None               Opt
	Restricted         Opt
	Valid              Opt
	RestrictedRegional Opt
	PowerSaving        Opt
}{
	func(id int) Opt { return domain.Resolve(id) },

	domain[0], domain[1],
	domain[2], domain[3], domain[4],
}

var roaming = optMap{
	0: Opt{0, "Non roaming"},
	1: Opt{1, "Roaming"},
}

// RoamingStates represent the state of roaming.
var RoamingStates = struct {
	Resolve func(int) Opt

	NotRoaming Opt
	Roaming    Opt
}{
	func(id int) Opt { return roaming.Resolve(id) },

	roaming[0], roaming[1],
}

var mode = optMap{
	0:  Opt{0, "No service"},
	1:  Opt{1, "AMPS"},
	2:  Opt{2, "CDMA"},
	3:  Opt{3, "GSM/GPRS"},
	4:  Opt{4, "HDR"},
	5:  Opt{5, "WCDMA"},
	6:  Opt{6, "GPS"},
	7:  Opt{7, "GSM/WCDMA"},
	8:  Opt{8, "CDMA/HDR HYBRID"},
	15: Opt{15, "TD-SCDMA"},
}

// SystemModes represent the possible system operating modes.
var SystemModes = struct {
	Resolve func(int) Opt

	NoService Opt
	AMPS      Opt
	CDMA      Opt
	GsmGprs   Opt
	HDR       Opt
	WCDMA     Opt
	GPS       Opt
	GsmWcdma  Opt
	CdmaHdr   Opt
	SCDMA     Opt
}{
	func(id int) Opt { return mode.Resolve(id) },

	mode[0], mode[1], mode[2], mode[3], mode[4],
	mode[5], mode[6], mode[7], mode[8], mode[15],
}

var submode = optMap{
	0:  Opt{0, "No service"},
	1:  Opt{1, "GSM"},
	2:  Opt{2, "GPRS"},
	3:  Opt{3, "EDGE"},
	4:  Opt{4, "WCDMA"},
	5:  Opt{5, "HSDPA"},
	6:  Opt{6, "HSUPA"},
	7:  Opt{7, "HSDPA and HSUPA"},
	8:  Opt{8, "TD-SCDMA"},
	9:  Opt{9, "HSPA+"},
	17: Opt{17, "HSPA+(64QAM)"},
	18: Opt{18, "HSPA+(MIMO)"},
}

// SystemSubmodes represent the possible system operating submodes.
var SystemSubmodes = struct {
	Resolve func(int) Opt

	NoService  Opt
	GSM        Opt
	GPRS       Opt
	EDGE       Opt
	WCDMA      Opt
	HSDPA      Opt
	HSUPA      Opt
	HsdpaHsupa Opt
	SCDMA      Opt
	HspaPlus   Opt
	Hspa64QAM  Opt
	HspaMIMO   Opt
}{
	func(id int) Opt { return submode.Resolve(id) },

	submode[0], submode[1], submode[2], submode[3],
	submode[4], submode[5], submode[6], submode[7],
	submode[8], submode[9], submode[17], submode[18],
}

var result = stringOpts{
	{"AT", "Noop"},
	{"OK", "Success"},
	{"CONNECT", "Connect"},
	{"RING", "Ringing"},
	{"NO CARRIER", "No carrier"},
	{"ERROR", "Error"},
	{"NO DIALTONE", "No dialtone"},
	{"BUSY", "Busy"},
	{"NO ANSWER", "No answer"},
	{"+CME ERROR:", "CME Error"},
	{"+CMS ERROR:", "CMS Error"},
	{"COMMAND NOT SUPPORT", "Command is not supported"},
	{"TOO MANY PARAMETERS", "Too many parameters"},
	{"AT_KILL", "Timeout"},
}

// FinalResults represent the possible replies from a modem.
var FinalResults = struct {
	Resolve func(string) StringOpt

	Noop              StringOpt
	Ok                StringOpt
	Connect           StringOpt
	Ring              StringOpt
	NoCarrier         StringOpt
	Error             StringOpt
	NoDialtone        StringOpt
	Busy              StringOpt
	NoAnswer          StringOpt
	CmeError          StringOpt
	CmsError          StringOpt
	NotSupported      StringOpt
	TooManyParameters StringOpt
	Timeout           StringOpt
}{
	func(str string) StringOpt { return result.Resolve(str) },

	result[0], result[1], result[2], result[3],
	result[4], result[5], result[6], result[7],
	result[8], result[9], result[10], result[11],
	result[12], result[13],
}

var resultReporting = optMap{
	0: Opt{0, "Disabled"},
	1: Opt{1, "Enabled"},
	2: Opt{2, "Exit"},
}

// UssdResultReporting represents the available options of USSD reporting.
var UssdResultReporting = struct {
	Resolve func(int) Opt

	Disable Opt
	Enable  Opt
	Exit    Opt
}{
	func(id int) Opt { return resultReporting.Resolve(id) },

	resultReporting[0],
	resultReporting[1],
	resultReporting[2],
}

var reports = stringOpts{
	{"+CUSD:", "USSD reply"},
	{"+CMTI:", "Incoming SMS"},
	{"^RSSI:", "Signal strength"},
	{"^BOOT:", "Boot handshake"},
	{"^MODE:", "System mode"},
	{"^SRVST:", "Service state"},
	{"^SIMST:", "Sim state"},
	{"^STIN:", "STIN"},
}

// Reports represent the possible state reports from a modem.
var Reports = struct {
	Resolve func(string) StringOpt

	Ussd           StringOpt
	Message        StringOpt
	SignalStrength StringOpt
	BootHandshake  StringOpt
	Mode           StringOpt
	ServiceState   StringOpt
	SimState       StringOpt
	Stin           StringOpt
}{
	func(str string) StringOpt { return reports.Resolve(str) },

	reports[0], reports[1], reports[2], reports[3],
	reports[4], reports[5], reports[6], reports[7],
}

var mem = stringOpts{
	{"ME", "NV RAM"},
	{"MT", "ME-associated storage"},
	{"SM", "Sim message storage"},
	{"SR", "State report storage"},
}

// MemoryTypes represent the available options of message storage.
var MemoryTypes = struct {
	Resolve func(string) StringOpt

	NvRAM       StringOpt
	Associated  StringOpt
	Sim         StringOpt
	StateReport StringOpt
}{
	func(str string) StringOpt { return mem.Resolve(str) },

	mem[0], mem[1], mem[2], mem[3],
}

var delOpts = optMap{
	0: Opt{0, "Delete message by index"},
	1: Opt{1, "Delete all read messages except MO"},
	2: Opt{2, "Delete all read messages except unsent MO"},
	3: Opt{3, "Delete all except unread"},
	4: Opt{4, "Delete all messages"},
}

// DeleteOptions represent the available options of message deletion masks.
var DeleteOptions = struct {
	Resolve func(int) Opt

	Index            Opt
	AllReadNotMO     Opt
	AllReadNotUnsent Opt
	AllNotUnread     Opt
	All              Opt
}{
	func(id int) Opt { return resultReporting.Resolve(id) },

	delOpts[0], delOpts[1], delOpts[2], delOpts[3], delOpts[4],
}

var msgFlags = optMap{
	0: Opt{0, "Unread"},
	1: Opt{1, "Read"},
	2: Opt{2, "Unsent"},
	3: Opt{3, "Sent"},
	4: Opt{4, "Any"},
}

// MessageFlags represent the available states of messages in memory.
var MessageFlags = struct {
	Resolve func(int) Opt

	Unread Opt
	Read   Opt
	Unsent Opt
	Sent   Opt
	Any    Opt
}{
	func(id int) Opt { return resultReporting.Resolve(id) },

	msgFlags[0], msgFlags[1], msgFlags[2], msgFlags[3], msgFlags[4],
}
