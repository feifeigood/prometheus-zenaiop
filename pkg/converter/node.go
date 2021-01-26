package converter

import (
	"errors"
	"fmt"

	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/prometheus/alertmanager/template"
	"go.uber.org/zap"
)

type nodeAlert struct {
	next Converter
}

func (na *nodeAlert) SetNext(next Converter) {
	na.next = next
}

func (na *nodeAlert) Convert(alerts *AIOPAlerts, wm webhook.Message) error {

	if alerts == nil {
		return errors.New("Alerting slice should not be nil")
	}

	as := template.Alerts{}
	for _, a := range wm.Alerts {
		a := a
		if !na.match(a.Labels) {
			as = append(as, a)
		} else {
			zap.S().Debugf("Source alerting(ECN-CDN-NODE) =>  %s", outputJSON(a))
			aa := AIOPAlert{
				ID:      FormatAIOPID(a.Labels),
				Type:    "ECN-CDN-NODE",
				Level:   FormatAIOPLevel(a.Labels["severity"]),
				Time:    FormatAIOPTime(a.StartsAt),
				Message: fmt.Sprintf("[%s] => %s", a.Labels["alertname"], a.Annotations["description"]),
				Infor:   fmt.Sprintf("ECN-CDN-NODE(%s)", a.Labels["address"]),
				Status:  FormatAIOPStatus(a.Status),
			}
			zap.S().Debugf("Target alerting(ECN-CDN-NODE) =>  %s", outputJSON(aa))
			*alerts = append(*alerts, aa)
		}
	}

	if len(as) != 0 && na.next != nil {
		return na.next.Convert(alerts, webhook.Message{
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

func (na *nodeAlert) match(labels template.KV) bool {

	// node alerting must contains 'alertname,address,'

	if _, ok := labels["address"]; !ok {
		return false
	}

	if _, ok := labels["alertname"]; !ok {
		return false
	}

	return true
}
