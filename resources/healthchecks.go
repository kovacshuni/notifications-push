package resources

import (
	"net/http"
	"time"

	"errors"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/kafka-client-go/kafka"
	"github.com/Financial-Times/service-status-go/gtg"
)

type HealthCheck struct {
	Consumer kafka.Consumer
}

func NewHealthCheck(kafkaConsumer kafka.Consumer) *HealthCheck {
	return &HealthCheck{
		Consumer: kafkaConsumer,
	}
}

func (h *HealthCheck) Health() func(w http.ResponseWriter, r *http.Request) {
	checks := []fthealth.Check{h.queueCheck()}
	hc := fthealth.TimedHealthCheck{
		HealthCheck: fthealth.HealthCheck{
			SystemCode:  "upp-notifications-push",
			Name:        "Notifications Push",
			Description: "Checks if all the dependent services are reachable and healthy.",
			Checks:      checks,
		},
		Timeout: 10 * time.Second,
	}
	return fthealth.Handler(hc)
}

// Check is the the NotificationsPushHealthcheck method that checks if the kafka queue is available
func (h *HealthCheck) queueCheck() fthealth.Check {
	return fthealth.Check{
		ID:               "message-queue-reachable",
		Name:             "MessageQueueReachable",
		Severity:         1,
		BusinessImpact:   "Notifications about newly modified/published content will not reach this app, nor will they reach its clients.",
		TechnicalSummary: "Message queue is not reachable/healthy",
		PanicGuide:       "https://dewey.ft.com/upp-notifications-push.html",
		Checker:          h.checkAggregateMessageQueueReachable,
	}
}

func (h *HealthCheck) GTG() gtg.Status {
	if _, err := h.checkAggregateMessageQueueReachable(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}

func (h *HealthCheck) checkAggregateMessageQueueReachable() (string, error) {
	// ISSUE: consumer's helthcheck always returns true
	err := h.Consumer.ConnectivityCheck()
	if err == nil {
		return "Connectivity to kafka is OK.", nil
	}

	return "Error connecting to kafka", errors.New("Error connecting to kafka queue")
}
