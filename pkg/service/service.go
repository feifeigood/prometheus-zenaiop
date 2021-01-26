package service

import (
	"encoding/json"
	"fmt"

	"github.com/feifeigood/prometheus-zenaiop/pkg/converter"
	"github.com/go-resty/resty/v2"
	"github.com/prometheus/alertmanager/notify/webhook"
	"go.uber.org/zap"
)

// PostResponse is the prometheus msteams service response.
type PostResponse struct {
	WebhookURL string `json:"webhook_url"`
	Status     int    `json:"status"`
	Message    string `json:"message"`
}

// Service is Alertmanager to Zenlayer AIOP webhook service.
type Service interface {
	Post(webhook.Message) (resp []PostResponse, err error)
}

type simpleService struct {
	converter  converter.Converter
	client     *resty.Client
	webhookURL string
}

// NewSimpleService creates a simpleService.
func NewSimpleService(converter converter.Converter, webhookURL string) Service {
	return simpleService{converter: converter, client: resty.New(), webhookURL: webhookURL}
}

func (s simpleService) Post(wm webhook.Message) ([]PostResponse, error) {
	alerts := converter.AIOPAlerts{}
	err := s.converter.Convert(&alerts, wm)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook message: %w", err)
	}

	resp, err := s.client.R().EnableTrace().SetHeader("Content-Type", "application/json").SetBody(jsonMarshal(map[string]interface{}{"alerts": alerts})).Post(s.webhookURL)
	if err != nil {
		return nil, err
	}

	zap.S().Infof("send notification to aiop webhook status: %d, body: %s", resp.StatusCode(), resp.Body())

	return nil, nil
}

func jsonMarshal(v interface{}) string {
	buf, _ := json.Marshal(v)
	return string(buf)
}
