package audit

import (
	"time"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/pkg/util"
	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

type URLObserver struct {
	url    string
	log    *zap.Logger
	client *resty.Client
}

func NewURLObserver(url string, log *zap.Logger) *URLObserver {
	logger := log.With(zap.String("[audit-url]", url))
	restyClient := resty.New().SetBaseURL(url).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetRetryCount(3).
		SetRetryWaitTime(time.Second)

	return &URLObserver{url: url, log: logger, client: restyClient}
}

func (o *URLObserver) Notify(message domain.AuditMessage) {
	if err := o.sendToURL(message); err != nil {
		o.log.Error("failed to send audit message to URL", zap.Error(err))
	}
	o.log.Info("audit message sent to URL", zap.String("url", o.url))
}

func (o *URLObserver) sendToURL(message domain.AuditMessage) error {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	compressedData, err := util.CompressJSON(jsonMessage)
	if err != nil {
		return err
	}

	_, err = o.client.R().SetBody(compressedData).Post(o.url)
	if err != nil {
		return err
	}

	return nil
}
