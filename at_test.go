package at

import (
	"log"
	"time"
)

// Needs to be changed for each particular configuration.
const (
	CommandPortPath  = "/dev/tty.HUAWEIMobile-Modem"
	NotifyPortPath   = "/dev/tty.HUAWEIMobile-Pcui"
	TestPhoneAddress = "+79269965690"
	BalanceUSSD      = "*100#"
)

var dev *Device

// openDevice opens the hardcoded device paths for reading and writing,
// also inits this device with the default device profile.
func openDevice() (err error) {
	dev = &Device{
		CommandPort: CommandPortPath,
		NotifyPort:  NotifyPortPath,
	}
	if err = dev.Open(); err != nil {
		return
	}
	if err = dev.Init(DeviceE173()); err != nil {
		return
	}
	return
}

// waitDevice monitors available channels for the given period of time, or
// until the fetch process exits.
func waitDevice(n int) {
	t := time.NewTimer(time.Second * time.Duration(n))
	defer t.Stop()
	go dev.Watch()
	for {
		select {
		case <-t.C:
			return
		case <-dev.Closed():
			return
		case msg, ok := <-dev.IncomingSms():
			if ok {
				log.Printf("Incoming sms from %s: %s", msg.Address, msg.Text)
			}
		case ussd, ok := <-dev.UssdReply():
			if ok {
				log.Printf("USSD result: %s", ussd)
			}
		case <-dev.StateUpdate():
			log.Printf("Signal strength: %d (%s/%s)", dev.State.SignalStrength, dev.State.OperatorName,
				dev.State.SystemSubmode.Description)
		}
	}
}

// This costs money (but works)

// func TestSmsSend(t *testing.T) {
// 	err := openDevice()
// 	if !assert.NoError(t, err) {
// 		return
// 	}
// 	defer dev.Close()
// 	msg := sms.Message{
// 		Text:     "Lazy fox jumps over ленивая собака",
// 		Type:     sms.MessageTypes.Submit,
// 		Encoding: sms.Encodings.UCS2,
// 		Address:  sms.PhoneNumber(TestPhoneAddress),
// 		VPFormat: sms.ValidityPeriodFormats.Relative,
// 		VP:       sms.ValidityPeriod(24 * time.Hour * 4),
// 	}
// 	n, octets, err := msg.PDU()
// 	if !assert.NoError(t, err) {
// 		return
// 	}

// 	err = dev.Commands.CMGS(n, octets)
// 	if !assert.NoError(t, err) {
// 		return
// 	}
// 	waitDevice(10)
// }
