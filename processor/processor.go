package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/flashbots/prometheus-sns-lambda-webhook/config"
	"github.com/flashbots/prometheus-sns-lambda-webhook/types"
	"go.uber.org/zap"
)

type Processor struct {
	includeBody bool
	log         *zap.Logger
	method      string
	url         string
}

func New(cfg *config.Config) (*Processor, error) {
	return &Processor{
		includeBody: cfg.Webhook.IncludeBody,
		log:         zap.L(),
		method:      cfg.Webhook.Method,
		url:         cfg.Webhook.Url,
	}, nil
}

func (p *Processor) processAlert(
	ctx context.Context,
	_ string,
	alert *types.Alert,
) error {
	var body io.Reader = nil
	if p.includeBody {
		body := new(bytes.Buffer)
		if err := json.NewEncoder(body).Encode(alert); err != nil {
			return err
		}
	}

	req, err := http.NewRequest(p.method, p.url, body)
	if err != nil {
		return err
	}
	if p.includeBody {
		req.Header.Set("content-type", "application/json")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if _, err := io.ReadAll(res.Body); err != nil {
		return err
	}

	return nil
}

func (p *Processor) processMessage(
	ctx context.Context,
	topic string,
	message *types.Message,
) error {
	errs := []error{}
	for _, alert := range message.Alerts {
		for k, v := range message.CommonAnnotations {
			if _, present := alert.Annotations[k]; !present {
				alert.Annotations[k] = v
			}
		}
		for k, v := range message.CommonLabels {
			if _, present := alert.Labels[k]; !present {
				alert.Labels[k] = v
			}
		}
		_timestamp, err := time.Parse("2006-01-02T15:04:05Z07:00", alert.StartsAt) // Grafana
		if err != nil {
			_timestamp, err = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", alert.StartsAt) // Prometheus
		}
		if err == nil {
			alert.StartsAt = _timestamp.Format("2006-01-02T15:04:05Z07:00")
		}
		if err := p.processAlert(ctx, topic, &alert); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (p *Processor) Lambda(ctx context.Context, event events.SNSEvent) error {
	l := p.log
	defer l.Sync() //nolint:errcheck

	errs := []error{}
	for _, r := range event.Records {
		var m types.Message
		if err := json.Unmarshal([]byte(r.SNS.Message), &m); err != nil {
			l.Error("Error un-marshalling message",
				zap.String("message", strings.Replace(r.SNS.Message, "\n", " ", -1)),
				zap.Error(err),
			)
			errs = append(errs, err)
			continue
		}
		if err := p.processMessage(ctx, r.SNS.TopicArn, &m); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}
