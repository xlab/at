package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/xlab/at"
	"github.com/xlab/at/sms"
)

func (m *Monitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Mon  *Monitor
		Dev  *at.Device
		Time time.Time
	}{
		Mon:  m,
		Dev:  m.dev,
		Time: time.Now(),
	}

	var buf bytes.Buffer
	err := tpl.Execute(&buf, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	io.Copy(w, &buf)
}

func decorateSignalStrength(n int) string {
	switch {
	case n == 0:
		return "< -113 dBm"
	case n == 1:
		return "-111 dBm"
	case n >= 2 && n <= 30:
		return fmt.Sprintf("%d dBm", -49-2*n)
	case n == 31:
		return "> -51 dBm"
	default:
		return "-"
	}
}

func decorateTime(t time.Time) string {
	return t.Format(time.RFC1123)
}

func decorateTimestamp(t sms.Timestamp) string {
	return time.Time(t).Format(time.RFC1123)
}

func inc(i int) int {
	return i + 1
}

var fm = template.FuncMap{
	"time":           decorateTime,
	"timestamp":      decorateTimestamp,
	"signalStrength": decorateSignalStrength,
	"inc":            inc,
}

var tpl = template.Must(template.New("index.html").Funcs(fm).Parse(indexTpl))

const indexTpl = `<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="utf-8">
    <title>{{ with .Dev.State }}{{ .ModelName }} {{ end }}Status</title>
    <link rel="stylesheet" href="http://maxcdn.bootstrapcdn.com/bootstrap/3.2.0/css/bootstrap.min.css">
</head>

<body>
    <div class="container">
        <div class="page-header">
            <h3>Modem status monitor
                <small>{{ .Dev.NotifyPort }}{{ with .Dev.State }} —{{ .IMEI }}{{ end }}</small>
            </h3>
        </div>
        <div class="row">
        	{{ if .Mon.Ready }}
            <div class="col-xs-6">
                <h4>Operator</h4>
                <p>{{ .Dev.State.OperatorName }}</p>
                <h4>Signal strength</h4>
                <p>{{ signalStrength .Dev.State.SignalStrength }}</p>
                <h4>Network mode</h4>
                <p>{{ .Dev.State.SystemSubmode.Description }}</p>
            </div>
            <div class="col-xs-6">
                <h4>Balance</h4>
                <p>{{with .Mon.Balance}}{{.}}{{ else }}-{{end}}</p>
                <h4>Received messages</h4>
                <p>{{ len .Mon.Messages }}</p>
                <h4>Last update</h4>
                <p>{{ time .Time }}</p>
            </div>
			{{ else }}
            <div class="col-xs-6">
                <h4>Status</h4>
                <p>Disconnected</p>
            </div>
            <div class="col-xs-6">
                <h4>Last update</h4>
                <p>{{ time .Time }}</p>
            </div>
            {{ end }}
        </div>
        <h3><i class="glyphicon glyphicon-inbox"></i> Inbox</h3>
        <table class="table">
        	{{ range $k,$v := .Mon.Messages }}
        	<tr>
                <th>{{ inc $k }}</th>
                <td>{{ timestamp $v.ServiceCenterTime }}</td>
                <td>{{ $v.Address }}</td>
                <td>{{ $v.Text }}</td>
            </tr>
            {{ else }}
            <tr>
                <td colspan="4">
                    <p class="text-muted">Empty</p>
                </td>
            </tr>
            {{ end }}
        </table>
    </div>
</body>

</html>`
