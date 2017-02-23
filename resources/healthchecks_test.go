package resources

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/stretchr/testify/assert"
)

var consumerConfig = consumer.QueueConfig{
	Group:            "push-group",
	Topic:            "content-notifications",
	AuthorizationKey: "my-first-auth-key",
}

func setupMockKafka(t *testing.T, status int, response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if status != 200 {
			w.WriteHeader(status)
		} else {
			w.Write([]byte(response))
		}

		assert.Equal(t, "my-first-auth-key", req.Header.Get("Authorization"))
	}))
}

func TestHealthchecks(t *testing.T) {
	kafka := setupMockKafka(t, 200, `["content-notifications"]`)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfig.Addrs = []string{kafka.URL}
	hc := NewNotificationsPushHealthcheck(consumerConfig)
	fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.Check())(w, req)

	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	t.Log(len(result.Checks))
	check := result.Checks[0]
	assert.True(t, check.BusinessImpact != "")
	assert.Equal(t, "MessageQueueProxyReachable", check.Name)
	assert.True(t, check.Ok)
	assert.True(t, result.Ok)
	assert.True(t, check.PanicGuide != "")
	assert.Equal(t, uint8(1), check.Severity)
}

func TestTopicMissing(t *testing.T) {
	kafka := setupMockKafka(t, 200, `[]`)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfig.Addrs = []string{kafka.URL}
	hc := NewNotificationsPushHealthcheck(consumerConfig)
	fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.Check())(w, req)

	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	check := result.Checks[0]
	assert.False(t, result.Ok)
	assert.False(t, check.Ok)
}

func TestTopicsUnparseable(t *testing.T) {
	kafka := setupMockKafka(t, 200, ``)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfig.Addrs = []string{kafka.URL}
	hc := NewNotificationsPushHealthcheck(consumerConfig)
	fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.Check())(w, req)

	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	check := result.Checks[0]
	assert.False(t, result.Ok)
	assert.False(t, check.Ok)
}

func TestFailingKafka(t *testing.T) {
	kafka := setupMockKafka(t, 500, ``)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfig.Addrs = []string{kafka.URL}
	hc := NewNotificationsPushHealthcheck(consumerConfig)
	fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.Check())(w, req)

	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	check := result.Checks[0]
	assert.False(t, result.Ok)
	assert.False(t, check.Ok)
}

func TestNoKafka(t *testing.T) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfig.Addrs = []string{"a-fake-url"}
	hc := NewNotificationsPushHealthcheck(consumerConfig)
	fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.Check())(w, req)

	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	check := result.Checks[0]
	assert.False(t, result.Ok)
	assert.False(t, check.Ok)
}

func TestGTG(t *testing.T) {
	kafka := setupMockKafka(t, 200, `["content-notifications"]`)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__gtg", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfig.Addrs = []string{kafka.URL}
	hc := NewNotificationsPushHealthcheck(consumerConfig)
	hc.GTG(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestGTGFailing(t *testing.T) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__gtg", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfig.Addrs = []string{"a-fake-url"}
	hc := NewNotificationsPushHealthcheck(consumerConfig)
	hc.GTG(w, req)

	assert.Equal(t, 503, w.Code)
}
