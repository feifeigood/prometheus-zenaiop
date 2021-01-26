package aiop

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/prometheus/alertmanager/notify/webhook"
	"go.uber.org/zap"
)

// Alert represents Zenlayer AIOP
// http://10.64.13.30:5006/api/v1.0/recive_alert/inter_alarm/SRE
type Alert struct {
	// ID is identifier for device using int type
	ID int64 `json:"id"`
	// Level is message priority level, the level order is 1 < 2 < 3
	Level int `json:"level"`
	// Type is message name for alert
	Type string `json:"type"`
	// Message is deatil for aiop
	Message string `json:"message"`
	// Infor represents resource of report message
	Infor string `json:"infor"`
	// Time message startAt, date format is 2020.08.07 15:22:12
	Time string `json:"time"`
	// Status message current status  PROBLEM/RESOLVED
	Status string `json:"status"`
	//Owner the people on call
	Owner string `json:"owner"`
}

// Converter converts an alert manager webhook message to AIOP
type Converter interface {
	Convert(webhook.Message) ([]Alert, error)
	// SetNext(Converter)
}

type aiopAlert struct {
}

// NewAIOPAlertCreator creates an AIOPAlert object.
func NewAIOPAlertCreator() Converter {
	return &aiopAlert{}
}

func (m *aiopAlert) Convert(wm webhook.Message) ([]Alert, error) {
	alerts := make([]Alert, 0)
	loc, _ := time.LoadLocation("Asia/Shanghai")

	for _, a := range wm.Alerts {
		t := time.Time(a.StartsAt)
		t = t.In(loc)

		alerts = append(alerts, Alert{
			Time: fmt.Sprintf("%d.%02d.%02d.%02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()),
			// custome labels the alert message must be contains
			ID:      ipv4toint(net.ParseIP(a.Labels["address"])),
			Level:   severitytolevel(a.Labels["severity"]),
			Type:    a.Labels["alertname"],
			Message: a.Annotations["description"],
			Infor:   a.Labels["address"],
			Status:  promstatustoaiop(a.Status),
			Owner:   a.Labels["owner"],
		})
	}

	zap.S().Infof("convert alertmanager message to aiop before: %s", jsonMarshal(wm))
	zap.S().Infof("convert alertmanager message to aiop after: %s", jsonMarshal(alerts))

	return alerts, nil
}

func jsonMarshal(v interface{}) string {
	buf, _ := json.Marshal(v)
	return string(buf)
}

func ipv4toint(ipv4 net.IP) int64 {
	if len(ipv4) == 16 {
		return int64(binary.BigEndian.Uint32(ipv4[12:16]))
	}
	return int64(binary.BigEndian.Uint32(ipv4))
}

func severitytolevel(severity string) int {
	switch severity {
	case "emergency":
		return 3
	case "critical":
		return 2
	default: // warning info or other using lowest level
		return 1
	}
}

func promstatustoaiop(status string) string {
	if status == "firing" {
		return "PROBLEM"
	}
	return "RESOLVED"
}
