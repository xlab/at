// Package at is a framework for communication with AT-compatible devices like Huawei modems via serial port.
// Currently this package is well-suited for Huawei devices and since AT-commands set may vary from device
// to device, sometimes you'll be forced to implement some logic by yourself.
//
// Framework
//
// This framework includes facilities for device monitoring, sending and receiving AT-commands, encoding and decoding SMS
// messages from or to PDU octet representation (as specified in 3GPP TS 23.040). An example of incoming SMS monitor application is given.
//
// Device-specific config
//
// In order to introduce your own logic (i.e. custom modem Init function), you should derive your profile from
// the default DeviceProfile and override its methods.
//
// About
//
// Project page: https://github.com/xlab/at
package at
