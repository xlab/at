package at

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/xlab/at/pdu"
	"github.com/xlab/at/sms"
	"github.com/xlab/at/util"
)

// DeviceProfile hides the device-specific implementation
// and provides a set of methods that can be used on a device.
// Init should be called first.
type DeviceProfile interface {
	Init(*Device) error
	CMGS(length int, octets []byte) (err error)
	CUSD(reporting Opt, octets []byte, enc Encoding) (err error)
	CMGR(index uint64) (octets []byte, err error)
	CMGD(index uint64, option Opt) (err error)
	CMGL(flag Opt) (octets map[uint64][]byte, err error)
	CMGF(text bool) (err error)
	CNMI(mode, mt, bm, ds, bfr int) (err error)
	CPMS(mem1 StringOpt, mem2 StringOpt, mem3 StringOpt) (err error)
	BOOT(token uint64) (err error)
	SYSCFG(roaming, cellular bool) (err error)
	SYSINFO() (info *SystemInfoReport, err error)
	COPS(auto bool, text bool) (err error)
	OperatorName() (str string, err error)
	ModelName() (str string, err error)
	IMEI() (str string, err error)
}

// DeviceE173 returns an instance of DeviceProfile implementation for Huawei E173,
// it's also the default one.
func DeviceE173() DeviceProfile {
	return &DefaultProfile{}
}

// DefaultProfile is a reference implementation that could be embedded
// in any other custom implementation of the DeviceProfile interface.
type DefaultProfile struct {
	dev *Device
}

// Init invokes a set of methods that will make the initial setup of the modem.
func (p *DefaultProfile) Init(d *Device) (err error) {
	p.dev = d
	p.dev.Send(NoopCmd) // kinda flush
	if err = p.COPS(true, true); err != nil {
		return errors.New("at init: unable to adjust the format of operator's name")
	}
	var info *SystemInfoReport
	if info, err = p.SYSINFO(); err != nil {
		return errors.New("at init: unable to read system info")
	}
	p.dev.State = &DeviceState{
		ServiceState:  info.ServiceState,
		ServiceDomain: info.ServiceDomain,
		RoamingState:  info.RoamingState,
		SystemMode:    info.SystemMode,
		SystemSubmode: info.SystemSubmode,
		SimState:      info.SimState,
	}
	if p.dev.State.OperatorName, err = p.OperatorName(); err != nil {
		return errors.New("at init: unable to read operator's name")
	}
	if p.dev.State.ModelName, err = p.ModelName(); err != nil {
		return errors.New("at init: unable to read modem's model name")
	}
	if p.dev.State.IMEI, err = p.IMEI(); err != nil {
		return errors.New("at init: unable to read modem's IMEI code")
	}
	if err = p.CMGF(false); err != nil {
		return errors.New("at init: unable to switch message format to PDU")
	}
	if err = p.CPMS(MemoryTypes.NvRAM, MemoryTypes.NvRAM, MemoryTypes.NvRAM); err != nil {
		return errors.New("at init: unable to set messages storage")
	}
	if err = p.CNMI(1, 1, 0, 0, 0); err != nil {
		return errors.New("at init: unable to turn on message notifications")
	}
	var octets map[uint64][]byte
	if octets, err = p.CMGL(MessageFlags.Any); err != nil {
		return errors.New("at init: unable to check message inbox")
	}
	for n, oct := range octets {
		var msg sms.Message
		if _, err := msg.ReadFrom(oct); err != nil {
			return errors.New("at init: error while parsing message inbox")
		}
		if err := p.CMGD(n, DeleteOptions.Index); err != nil {
			return errors.New("at init: error while cleaning message inbox")
		}
		d.messages <- &msg
	}
	return nil
}

type signalStrengthReport uint64

func (s *signalStrengthReport) Parse(str string) (err error) {
	var u uint64
	u, err = strconv.ParseUint(str, 10, 8)
	*s = signalStrengthReport(u)
	return
}

type modeReport struct {
	Mode    Opt
	Submode Opt
}

func (m *modeReport) Parse(str string) (err error) {
	fields := strings.Split(str, ",")
	if len(fields) < 2 {
		return ErrParseReport
	}
	var mode, submode uint64
	if mode, err = strconv.ParseUint(fields[0], 10, 8); err != nil {
		return
	}
	if submode, err = strconv.ParseUint(fields[1], 10, 8); err != nil {
		return
	}
	m.Mode = SystemModes.Resolve(int(mode))
	m.Submode = SystemSubmodes.Resolve(int(submode))
	return
}

type simStateReport Opt

func (s *simStateReport) Parse(str string) (err error) {
	var o uint64
	if o, err = strconv.ParseUint(str, 10, 8); err != nil {
		return
	}
	*s = simStateReport(SimStates.Resolve(int(o)))
	return
}

type serviceStateReport Opt

func (s *serviceStateReport) Parse(str string) (err error) {
	var o uint64
	if o, err = strconv.ParseUint(str, 10, 8); err != nil {
		return
	}
	*s = serviceStateReport(ServiceStates.Resolve(int(o)))
	return
}

type bootHandshakeReport uint64

func (b *bootHandshakeReport) Parse(str string) (err error) {
	fields := strings.Split(str, ",")
	if len(fields) < 1 {
		return ErrParseReport
	}
	var key uint64
	if key, err = strconv.ParseUint(fields[0], 10, 8); err != nil {
		return
	}
	*b = bootHandshakeReport(key)
	return
}

// Ussd type represents the USSD query string.
type Ussd string

// Encode converts the query string into bytes according to the
// specified encoding.
func (u *Ussd) Encode(enc Encoding) ([]byte, error) {
	switch enc {
	case Encodings.Gsm7Bit:
		return pdu.Encode7Bit(u.String()), nil
	case Encodings.UCS2:
		return pdu.EncodeUcs2(u.String()), nil
	default:
		return nil, ErrUnknownEncoding
	}
}

func (u *Ussd) String() string {
	return string(*u)
}

type ussdReport struct {
	N      uint64
	Octets []byte
	Enc    Encoding
}

func (r *ussdReport) Parse(str string) (err error) {
	fields := strings.Split(str, ",")
	if len(fields) < 3 {
		return ErrParseReport
	}
	if r.N, err = strconv.ParseUint(fields[0], 10, 8); err != nil {
		return
	}
	if r.Octets, err = util.Bytes(strings.Trim(fields[1], `"`)); err != nil {
		return
	}
	var e uint64
	if e, err = strconv.ParseUint(fields[2], 10, 8); err != nil {
		return
	}
	r.Enc = Encoding(e)
	return
}

// CUSD sends AT+CUSD with the given parameters to the device. This will invoke an USSD request.
func (p *DefaultProfile) CUSD(reporting Opt, octets []byte, enc Encoding) (err error) {
	req := fmt.Sprintf(`AT+CUSD=%d,%02X,%d`, reporting.ID, octets, enc)
	_, err = p.dev.Send(req)
	return
}

type messageReport struct {
	Memory StringOpt
	Index  uint64
}

func (m *messageReport) Parse(str string) (err error) {
	fields := strings.Split(str, ",")
	if len(fields) < 2 {
		return ErrParseReport
	}
	if m.Memory = MemoryTypes.Resolve(strings.Trim(fields[0], `"`)); m.Memory == UnknownStringOpt {
		return ErrParseReport
	}
	if m.Index, err = strconv.ParseUint(fields[1], 10, 16); err != nil {
		return
	}
	return
}

// CMGR sends AT+CMGR with the given index to the device and returns the message contents.
func (p *DefaultProfile) CMGR(index uint64) (octets []byte, err error) {
	req := fmt.Sprintf(`AT+CMGR=%d`, index)
	reply, err := p.dev.Send(req)
	if err != nil {
		return
	}
	lines := strings.Split(reply, "\n")
	if len(lines) < 2 {
		return nil, ErrParseReport
	}
	octets, err = util.Bytes(lines[1])
	return
}

// CMGD sends AT+CMGD with the given index and option to the device. Option defines the mode
// in which messages will be deleted. The default mode is to delete by index.
func (p *DefaultProfile) CMGD(index uint64, option Opt) (err error) {
	req := fmt.Sprintf(`AT+CMGD=%d,%d`, index, option.ID)
	_, err = p.dev.Send(req)
	return
}

// CPMS sends AT+CPMS with the given options to the device. It allows to select
// the storage type for different kinds of messages and message notifications.
func (p *DefaultProfile) CPMS(mem1 StringOpt, mem2 StringOpt, mem3 StringOpt) (err error) {
	req := fmt.Sprintf(`AT+CPMS="%s","%s","%s"`, mem1.ID, mem2.ID, mem3.ID)
	_, err = p.dev.Send(req)
	return
}

// CNMI sends AT+CNMI with the given parameters to the device.
// It's used to adjust the settings of the new message arrival notifications.
func (p *DefaultProfile) CNMI(mode, mt, bm, ds, bfr int) (err error) {
	req := fmt.Sprintf(`AT+CNMI=%d,%d,%d,%d,%d`, mode, mt, bm, ds, bfr)
	_, err = p.dev.Send(req)
	return
}

// CMGF sends AT+CMGF with the given value to the device. It toggles
// the mode of message handling between PDU and TEXT.
//
// Note, that the at package works only in PDU mode.
func (p *DefaultProfile) CMGF(text bool) (err error) {
	var flag int
	if text {
		flag = 1
	}
	req := fmt.Sprintf(`AT+CMGF=%d`, flag)
	_, err = p.dev.Send(req)
	return
}

// CMGL sends AT+CMGL with the given filtering flag to the device and then parses
// the list of received messages that match their filter. See MessageFlags for the
// list of supported filters.
func (p *DefaultProfile) CMGL(flag Opt) (octets map[uint64][]byte, err error) {
	req := fmt.Sprintf(`AT+CMGL=%d`, flag.ID)
	reply, err := p.dev.Send(req)
	if err != nil {
		return
	}
	lines := strings.Split(reply, "\n")
	if len(lines) < 2 {
		return
	}
	octets = make(map[uint64][]byte)
	for i := 0; i < len(lines); i += 2 {
		header := strings.TrimPrefix(lines[i], `+CMGL: `)
		fields := strings.Split(header, ",")
		if len(fields) < 4 {
			return nil, ErrParseReport
		}
		n, err := strconv.ParseUint(fields[0], 10, 16)
		if err != nil {
			return nil, ErrParseReport
		}
		var oct []byte
		if oct, err = util.Bytes(lines[i+1]); err != nil {
			return nil, ErrParseReport
		}
		octets[n] = oct
	}
	return
}

// BOOT sends AT^BOOT with the given token to the device. This completes
// the handshaking procedure.
func (p *DefaultProfile) BOOT(token uint64) (err error) {
	req := fmt.Sprintf(`AT^BOOT=%d,0`, token)
	_, err = p.dev.Send(req)
	return
}

// CMGS sends AT+CMGS with the given parameters to the device. This is used to send SMS
// using the given PDU data. Length is a number of TPDU bytes.
func (p *DefaultProfile) CMGS(length int, octets []byte) (err error) {
	part1 := fmt.Sprintf("AT+CMGS=%d", length)
	part2 := fmt.Sprintf("%02X", octets)
	err = p.dev.sendInteractive(part1, part2, byte('>'))
	return
}

// SYSCFG sends AT^SYSCFG with the given parameters to the device.
// The arguments of this command may vary, so the options are limited to switchng roaming and
// cellular mode on/off.
func (p *DefaultProfile) SYSCFG(roaming, cellular bool) (err error) {
	var roam int
	if roaming {
		roam = 1
	}
	var cell int
	if cellular {
		cell = 2
	} else {
		cell = 1
	}
	req := fmt.Sprintf(`AT^SYSCFG=2,2,3FFFFFFF,%d,%d`, roam, cell)
	_, err = p.dev.Send(req)
	return
}

// SystemInfoReport represents the report from the AT^SYSINFO command.
type SystemInfoReport struct {
	ServiceState  Opt
	ServiceDomain Opt
	RoamingState  Opt
	SystemMode    Opt
	SystemSubmode Opt
	SimState      Opt
}

// Parse scans the AT^SYSINFO report into a non-nil SystemInfoReport struct.
func (s *SystemInfoReport) Parse(str string) (err error) {
	fields := strings.Split(str, ",")
	if len(fields) < 7 {
		return ErrParseReport
	}

	fetch := func(str string, field *Opt, resolver func(id int) Opt) error {
		if n, err := strconv.ParseUint(str, 10, 8); err != nil {
			return err
		} else if opt := resolver(int(n)); opt == UnknownOpt {
			return errors.New("resolver: unknown opt")
		} else {
			*field = opt
			return nil
		}
	}

	if err = fetch(fields[0], &s.ServiceState, ServiceStates.Resolve); err != nil {
		return ErrParseReport
	}
	if err = fetch(fields[1], &s.ServiceDomain, ServiceDomains.Resolve); err != nil {
		return ErrParseReport
	}
	if err = fetch(fields[2], &s.RoamingState, RoamingStates.Resolve); err != nil {
		return ErrParseReport
	}
	if err = fetch(fields[3], &s.SystemMode, SystemModes.Resolve); err != nil {
		return ErrParseReport
	}
	if err = fetch(fields[4], &s.SimState, SimStates.Resolve); err != nil {
		return ErrParseReport
	}
	if err = fetch(fields[6], &s.SystemSubmode, SystemSubmodes.Resolve); err != nil {
		return ErrParseReport
	}
	return nil
}

// SYSINFO sends AT^SYSINFO to the device and parses the output.
func (p *DefaultProfile) SYSINFO() (info *SystemInfoReport, err error) {
	reply, err := p.dev.Send(`AT^SYSINFO`)
	if err != nil {
		return nil, err
	}
	info = new(SystemInfoReport)
	err = info.Parse(strings.TrimPrefix(reply, `^SYSINFO:`))
	return
}

// COPS sends AT+COPS to the device with parameters that define autosearch and
// the operator's name representation. The default representation is numerical.
func (p *DefaultProfile) COPS(auto bool, text bool) (err error) {
	var a, t int
	if !auto {
		a = 1
	}
	if !text {
		t = 2
	}
	req := fmt.Sprintf(`AT+COPS=%d,%d`, a, t)
	_, err = p.dev.Send(req)
	return
}

// OperatorName sends AT+COPS? to the device and gets the operator's name.
func (p *DefaultProfile) OperatorName() (str string, err error) {
	result, err := p.dev.Send(`AT+COPS?`)
	fields := strings.Split(strings.TrimPrefix(result, `+COPS: `), ",")
	if len(fields) < 4 {
		err = ErrParseReport
		return
	}
	str = strings.TrimLeft(strings.TrimRight(fields[2], `"`), `"`)
	return
}

// ModelName sends AT+GMM to the device and gets the modem's model name.
func (p *DefaultProfile) ModelName() (str string, err error) {
	str, err = p.dev.Send(`AT+GMM`)
	return
}

// IMEI sends AT+GSN to the device and gets the modem's IMEI code.
func (p *DefaultProfile) IMEI() (str string, err error) {
	str, err = p.dev.Send(`AT+GSN`)
	return
}
