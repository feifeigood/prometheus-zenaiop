package converter

import (
	"errors"

	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/prometheus/alertmanager/template"
	"go.uber.org/zap"
)

type unknowAlert struct {
	next Converter
}

func (ua *unknowAlert) SetNext(next Converter) {
	ua.next = next
}

func (ua *unknowAlert) Convert(alerts *AIOPAlerts, wm webhook.Message) error {
	if alerts == nil {
		return errors.New("Alerting slice should not be nil")
	}

	for _, a := range wm.Alerts {
		zap.S().Warnf("Unknow alerting(UNKNOW) =>  %s", outputJSON(a))
	}

	return nil
}

func (ua *unknowAlert) match(labels template.KV) bool {

	return true
}
