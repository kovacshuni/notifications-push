package resources

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/stretchr/testify/assert"
	"net/http"
)

func initializeHealthCheck(isConsumerConnectionHealthy bool) *HealthCheck {
	return &HealthCheck{
		consumer: &mockConsumerInstance{isConnectionHealthy: isConsumerConnectionHealthy},
	}
}

func TestNewHealthCheck(t *testing.T) {
	hc := NewHealthCheck(consumer.NewConsumer(consumer.QueueConfig{}, func(m consumer.Message) {}, http.DefaultClient))

	assert.NotNil(t, hc.consumer)
}

func TestHappyHealthCheck(t *testing.T) {
	hc := initializeHealthCheck(true)

	req := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	hc.Health()(w, req)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Message Queue Proxy Reachable","ok":true`, "Message queue proxy healthcheck should be happy")
}

func TestUnhappyHealthCheck(t *testing.T) {
	hc := initializeHealthCheck(false)

	req := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	hc.Health()(w, req)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Message Queue Proxy Reachable","ok":false`, "Message queue proxy healthcheck should be unhappy")
}

func TestGTGHappyFlow(t *testing.T) {
	hc := initializeHealthCheck(true)

	status := hc.GTG()
	assert.True(t, status.GoodToGo)
	assert.Empty(t, status.Message)
}

func TestGTGBrokenConsumer(t *testing.T) {
	hc := initializeHealthCheck(false)

	status := hc.GTG()
	assert.False(t, status.GoodToGo)
	assert.Equal(t, "Error connecting to the queue", status.Message)
}

type mockConsumerInstance struct {
	isConnectionHealthy bool
}

func (c *mockConsumerInstance) Start() {
}

func (c *mockConsumerInstance) Stop() {
}

func (c *mockConsumerInstance) ConnectivityCheck() (string, error) {
	if c.isConnectionHealthy {
		return "", nil
	}

	return "", errors.New("Error connecting to the queue")
}
