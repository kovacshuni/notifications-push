package resources

import (
	"errors"
	"net/http"

	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/kafka-client-go/kafka"
)

//NotificationsPushHealthcheck is the main health check for the notifications-push service
type NotificationsPushHealthcheck struct {
	Client   *http.Client
	Consumer kafka.Consumer
}

// NewNotificationsPushHealthcheck returns a new instance of NotificationsPushHealthcheck.
func NewNotificationsPushHealthcheck(kafkaConsumer kafka.Consumer) *NotificationsPushHealthcheck {
	return &NotificationsPushHealthcheck{
		Client:   &http.Client{},
		Consumer: kafkaConsumer,
	}
}

// GTG is the HTTP handler function for the good-to-go endpoint of notifications-push
func (hc *NotificationsPushHealthcheck) GTG(w http.ResponseWriter, req *http.Request) {
	if _, err := hc.checkAggregateMessageQueueReachable(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Kafka healthcheck failed"))
	}
}

// Check is the the NotificationsPushHealthcheck method that checks if the kafka queue is available
func (hc *NotificationsPushHealthcheck) Check() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Notifications about newly modified/published content will not reach this app, nor will they reach its clients.",
		Name:             "MessageQueueProxyReachable",
		PanicGuide:       "https://sites.google.com/a/ft.com/universal-publishing/ops-guides/notifications-push",
		Severity:         1,
		TechnicalSummary: "Message queue proxy is not reachable/healthy",
		Checker:          hc.checkAggregateMessageQueueReachable,
	}
}

func (hc *NotificationsPushHealthcheck) checkAggregateMessageQueueReachable() (string, error) {
	// ISSUE: consumer's helthcheck always returns true
	err := hc.Consumer.ConnectivityCheck()
	if err == nil {
		return "Connectivity to kafka is OK.", nil
	}

	return "Error connecting to kafka", errors.New("Error connecting to kafka queue")
}
