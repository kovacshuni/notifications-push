package resources

import (
	"errors"
	"net/http"

	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/wvanbergen/kafka/consumergroup"
)

//NotificationsPushHealthcheck is the main health check for the notifications-push service
type NotificationsPushHealthcheck struct {
	consumer *consumergroup.ConsumerGroup
}

// NewNotificationsPushHealthcheck returns a new instance of NotificationsPushHealthcheck.
// It requires the configuration of the consumer queue.
func NewNotificationsPushHealthcheck(consumer *consumergroup.ConsumerGroup) *NotificationsPushHealthcheck {
	return &NotificationsPushHealthcheck{
		consumer: consumer,
	}
}

// GTG is the HTTP handler function for the good-to-go endpoint of notifications-push
func (hc *NotificationsPushHealthcheck) GTG(w http.ResponseWriter, req *http.Request) {
	if _, err := hc.checkAggregateMessageQueueProxiesReachable(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

// Check is the the NotificationsPushHealthcheck method that checks is the queue proxies are available
func (hc *NotificationsPushHealthcheck) Check() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Notifications about newly modified/published content will not reach this app, nor will they reach its clients.",
		Name:             "MessageQueueProxyReachable",
		PanicGuide:       "https://sites.google.com/a/ft.com/universal-publishing/ops-guides/notifications-push",
		Severity:         1,
		TechnicalSummary: "Message queue proxy is not reachable/healthy",
		Checker:          hc.checkAggregateMessageQueueProxiesReachable,
	}
}

func (hc *NotificationsPushHealthcheck) checkAggregateMessageQueueProxiesReachable() (string, error) {
	_, err := hc.consumer.InstanceRegistered()

	if err == nil {
		return "Connectivity to proxies is OK.", nil
	}
	errMsg := err.Error()

	return "Error connecting to proxies", errors.New(errMsg)
}
