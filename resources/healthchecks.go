package resources

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/Financial-Times/go-fthealth"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
)

// HealthcheckConfig contains healthcheck related config (i.e. which queue we're using)
type HealthcheckConfig struct {
	Client         *http.Client
	ConsumerConfig consumer.QueueConfig
}

// Health returns the /__health endpoint
func Health(config HealthcheckConfig) func(w http.ResponseWriter, r *http.Request) {
	return fthealth.HandlerParallel("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", messageQueueProxyReachable(config))
}

// GTG returns the /__gtg endpoint
func GTG(config HealthcheckConfig) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		healthChecks := []func() error{checkAggregateMessageQueueProxiesReachable(config)}

		for _, hCheck := range healthChecks {
			if err := hCheck(); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}
	}
}

func messageQueueProxyReachable(config HealthcheckConfig) fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Notifications about newly modified/published content will not reach this app, nor will they reach its clients.",
		Name:             "MessageQueueProxyReachable",
		PanicGuide:       "https://sites.google.com/a/ft.com/technology/systems/dynamic-semantic-publishing/extra-publishing/notifications-push-runbook",
		Severity:         1,
		TechnicalSummary: "Message queue proxy is not reachable/healthy",
		Checker:          checkAggregateMessageQueueProxiesReachable(config),
	}
}

func checkAggregateMessageQueueProxiesReachable(config HealthcheckConfig) func() error {
	return func() error {
		errMsg := ""
		for _, address := range config.ConsumerConfig.Addrs {
			err := checkMessageQueueProxyReachable(config, address)
			if err == nil {
				return nil
			}
			errMsg = errMsg + fmt.Sprintf("For %s there is an error %v \n", address, err.Error())
		}
		return errors.New(errMsg)
	}
}

func checkMessageQueueProxyReachable(config HealthcheckConfig, address string) error {
	req, err := http.NewRequest("GET", address+"/topics", nil)
	if err != nil {
		log.Warnf("Could not connect to proxy: %v", err.Error())
		return err
	}

	if len(config.ConsumerConfig.AuthorizationKey) > 0 {
		req.Header.Add("Authorization", config.ConsumerConfig.AuthorizationKey)
	}

	if len(config.ConsumerConfig.Queue) > 0 {
		req.Host = config.ConsumerConfig.Queue
	}

	resp, err := config.Client.Do(req)
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
	return checkIfTopicIsPresent(body, config.ConsumerConfig.Topic)
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
