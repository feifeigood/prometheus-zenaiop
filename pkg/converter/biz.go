package converter

import (
	"errors"
	"fmt"

	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/prometheus/alertmanager/template"
	"go.uber.org/zap"
)

type bizAlert struct {
	next Converter
}

func (ba *bizAlert) SetNext(next Converter) {
	ba.next = next
}

func (ba *bizAlert) Convert(alerts *AIOPAlerts, wm webhook.Message) error {

	if alerts == nil {
		return errors.New("Alerting slice should not be nil")
	}

	as := template.Alerts{}
	for _, a := range wm.Alerts {
		a := a
		if !ba.match(a.Labels) {
			as = append(as, a)
		} else {
			zap.S().Debugf("Source alerting(ECN-CDN-BIZ) =>  %s", outputJSON(a))
			aa := AIOPAlert{
				ID:      FormatAIOPID(a.Labels),
				Type:    "ECN-CDN-BIZ",
				Level:   FormatAIOPLevel(a.Labels["severity"]),
				Time:    FormatAIOPTime(a.StartsAt),
				Message: fmt.Sprintf("[%s] => %s", a.Labels["alertname"], a.Annotations["description"]),
				Infor:   fmt.Sprintf("ECN-CDN-BIZ(%s)", a.Labels["domain"]),
				Status:  FormatAIOPStatus(a.Status),
			}
			zap.S().Debugf("Target alerting(ECN-CDN-BIZ) =>  %s", outputJSON(aa))
			*alerts = append(*alerts, aa)
		}
	}

	if len(as) != 0 && ba.next != nil {
		return ba.next.Convert(alerts, webhook.Message{
			Data: &template.Data{
				Receiver:          wm.Receiver,
				Status:            wm.Status,
				Alerts:            as,
				GroupLabels:       wm.GroupLabels,
				CommonLabels:      wm.CommonLabels,
				CommonAnnotations: wm.CommonAnnotations,
				ExternalURL:       wm.ExternalURL,
			},
			Version:         wm.Version,
			GroupKey:        wm.GroupKey,
			TruncatedAlerts: wm.TruncatedAlerts,
		})
	}

	return nil
}

func (ba *bizAlert) match(labels template.KV) bool {

	// node alerting must contains 'alertname,domain,'

	if _, ok := labels["domain"]; !ok {
		return false
	}

	if _, ok := labels["alertname"]; !ok {
		return false
	}

	return true
}
