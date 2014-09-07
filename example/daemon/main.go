package main

import "log"

const (
	CommandPortPath = "/dev/tty.HUAWEIMobile-Modem"
	NotifyPortPath  = "/dev/tty.HUAWEIMobile-Pcui"
	WebPort         = 5051
)

func main() {
	log.Printf("Staring daemon at http://localhost:%d", WebPort)
	mon := NewMonitor(CommandPortPath, NotifyPortPath)

	if err := mon.Run(); err != nil {
		log.Fatalln(err)
	}
}
