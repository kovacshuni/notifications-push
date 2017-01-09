package resources

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"

	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
)

//NotificationsPushHealthcheck is the main health check for the notifications-push service
type NotificationsPushHealthcheck struct {
	Client         *http.Client
	ConsumerConfig consumer.QueueConfig
}

// NewNotificationsPushHealthcheck returns a new instance of NotificationsPushHealthcheck.
// It requires the configuration of the consumer queue.
func NewNotificationsPushHealthcheck(consumerConfig consumer.QueueConfig) *NotificationsPushHealthcheck {
	return &NotificationsPushHealthcheck{
		Client:         &http.Client{},
		ConsumerConfig: consumerConfig,
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
	errMsg := ""
	for _, address := range hc.ConsumerConfig.Addrs {
		err := hc.checkMessageQueueProxyReachable(address)
		if err == nil {
			return "Connectivity to proxies is OK.", nil
		}
		errMsg = errMsg + fmt.Sprintf("For %s there is an error %v \n", address, err.Error())
	}
	return "Error connecting to proxies", errors.New(errMsg)
}

func (hc *NotificationsPushHealthcheck) checkMessageQueueProxyReachable(address string) error {
	req, err := http.NewRequest("GET", address+"/topics", nil)
	if err != nil {
		log.Warnf("Could not connect to proxy: %v", err.Error())
		return err
	}

	if len(hc.ConsumerConfig.AuthorizationKey) > 0 {
		req.Header.Add("Authorization", hc.ConsumerConfig.AuthorizationKey)
	}

	if len(hc.ConsumerConfig.Queue) > 0 {
		req.Host = hc.ConsumerConfig.Queue
	}

	resp, err := hc.Client.Do(req)
	if err != nil {
		log.Warnf("Could not connect to proxy: %v", err.Error())
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Proxy returned status: %d", resp.StatusCode)
		return errors.New(errMsg)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return checkIfTopicIsPresent(body, hc.ConsumerConfig.Topic)
}

func checkIfTopicIsPresent(body []byte, searchedTopic string) error {
	var topics []string

	err := json.Unmarshal(body, &topics)
	if err != nil {
		return fmt.Errorf("Error occured and topic could not be found. %v", err.Error())
	}

	for _, topic := range topics {
		if topic == searchedTopic {
			return nil
		}
	}

	return errors.New("Topic was not found")
}
