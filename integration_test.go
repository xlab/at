//go:build !integration
// +build !integration

package at

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xlab/at/pdu"
)

// Test the device lifecycle.
func TestOpenInitWaitClose(t *testing.T) {
	err := openDevice()
	if !assert.NoError(t, err) {
		return
	}
	defer dev.Close()
	waitDevice(1)
}

// Test the "AT" command.
func TestNoop(t *testing.T) {
	err := openDevice()
	if !assert.NoError(t, err) {
		return
	}
	defer dev.Close()
	_, err = dev.Send(NoopCmd)
	assert.NoError(t, err)
}

// Test USSD queries, the result will be reported asynchroniously.
func TestUssd(t *testing.T) {
	err := openDevice()
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		dev.Close()
	}()
	err = dev.Commands.CUSD(UssdResultReporting.Enable, pdu.Encode7Bit(BalanceUSSD), Encodings.Gsm7Bit)
	if !assert.NoError(t, err) {
		return
	}
	waitDevice(10)
}
