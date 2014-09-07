## AT / example / daemon

This utility should be used as an example of the 'at' package usage. The daemon waits for a modem device that is specified in constants.

```go
	CommandPortPath = "/dev/tty.HUAWEIMobile-Modem"
	NotifyPortPath  = "/dev/tty.HUAWEIMobile-Pcui"
```

And monitors its SMS inbox and balance.

```go
	BalanceUSSD          = "*100#"
	BalanceCheckInterval = time.Minute
	DeviceCheckInterval  = time.Second * 10
```

It also spawns a web interface available at `http://localhost:%d`.
