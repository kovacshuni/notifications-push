package resources

import (
	"net/http"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/service-status-go/gtg"
)

const requestTimeout = 4500

type HealthCheck struct {
	consumer consumer.MessageConsumer
}

func NewHealthCheck(consumerConf *consumer.QueueConfig) *HealthCheck {
	httpClient := &http.Client{Timeout: requestTimeout * time.Millisecond}
	c := consumer.NewConsumer(*consumerConf, func(m consumer.Message) {}, httpClient)
	return &HealthCheck{
		consumer: c,
	}
}

func (h *HealthCheck) Health() func(w http.ResponseWriter, r *http.Request) {
	checks := []fthealth.Check{h.queueCheck()}
	hc := fthealth.HealthCheck{
		SystemCode:  "upp-notifications-push",
		Name:        "Notifications Push",
		Description: "Checks if all the dependent services are reachable and healthy.",
		Checks:      checks,
	}
	return fthealth.Handler(hc)
}

func (h *HealthCheck) queueCheck() fthealth.Check {
	return fthealth.Check{
		ID:               "message-queue-proxy-reachable",
		Name:             "Message Queue Proxy Reachable",
		Severity:         1,
		BusinessImpact:   "Notifications about newly modified/published content will not reach this app, nor will they reach its clients.",
		TechnicalSummary: "Message queue proxy is not reachable/healthy",
		PanicGuide:       "https://dewey.ft.com/upp-notifications-push.html",
		Checker:          h.consumer.ConnectivityCheck,
	}
}

func (h *HealthCheck) GTG() gtg.Status {
	if _, err := h.consumer.ConnectivityCheck(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}
