package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/xlab/at"
	"github.com/xlab/at/sms"
)

const (
	BalanceUSSD          = "*100#"
	BalanceCheckInterval = time.Minute
	DeviceCheckInterval  = time.Second * 10
)

type State uint8

const (
	NoDeviceState State = iota
	ReadyState
)

type Monitor struct {
	// Messages is the most simpliest volatile DB storing sms messages.
	Messages []*sms.Message
	// Balance is the balance reply we've got with an USSD query.
	Balance string
	// Ready signals if device is ready.
	Ready bool

	cmdPort    string
	notifyPort string

	dev          *at.Device
	stateChanged chan State
	checkTimer   *time.Timer
}

func (m *Monitor) DeviceState() *at.DeviceState {
	return m.dev.State
}

func NewMonitor(cmdPort, notifyPort string) *Monitor {
	return &Monitor{
		cmdPort:      cmdPort,
		notifyPort:   notifyPort,
		stateChanged: make(chan State, 10),
	}
}

func (m *Monitor) devStop() {
	if m.dev != nil {
		m.dev.Close()
	}
}

func (m *Monitor) Run() (err error) {
	m.checkTimer = time.NewTimer(DeviceCheckInterval)
	defer m.checkTimer.Stop()
	defer m.devStop()

	go func() {
		for {
			<-m.checkTimer.C
			if err := m.openDevice(); err != nil {
				m.checkTimer.Reset(DeviceCheckInterval)
				continue
			} else {
				m.checkTimer.Stop()
				m.stateChanged <- ReadyState
			}
		}
	}()

	if err := m.openDevice(); err != nil {
		m.stateChanged <- NoDeviceState
	} else {
		m.stateChanged <- ReadyState
		m.checkTimer.Stop()
	}

	go func() {
		for s := range m.stateChanged {
			switch s {
			case NoDeviceState:
				m.Balance = ""
				m.Ready = false
				log.Println("Waiting for device")
				m.checkTimer.Reset(DeviceCheckInterval)
			case ReadyState:
				log.Println("Device connected")
				m.Ready = true
				go func() {
					m.dev.Watch()
					m.stateChanged <- NoDeviceState
				}()
				go func() {
					m.dev.SendUSSD(BalanceUSSD)
					t := time.NewTicker(BalanceCheckInterval)
					defer t.Stop()
					for {
						select {
						case <-m.dev.Closed():
							return
						case ussd, ok := <-m.dev.UssdReply():
							if ok {
								m.Balance = string(ussd)
							}
						case msg, ok := <-m.dev.IncomingSms():
							if ok {
								m.Messages = append(m.Messages, msg)
							}
						case <-t.C:
							m.dev.SendUSSD(BalanceUSSD)
						}
					}
				}()
			}
		}
	}()

	return http.ListenAndServe(":"+strconv.Itoa(WebPort), m)
}

func (m *Monitor) openDevice() (err error) {
	m.dev = &at.Device{
		CommandPort: m.cmdPort,
		NotifyPort:  m.notifyPort,
	}
	if err = m.dev.Open(); err != nil {
		return
	}
	if err = m.dev.Init(at.DeviceE173()); err != nil {
		return
	}
	return
}
