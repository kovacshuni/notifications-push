package mocks

import (
	"github.com/Financial-Times/notifications-push/dispatcher"
	"github.com/stretchr/testify/mock"
	"net/http"
	"io/ioutil"
	"strings"
	"errors"
)

// MockDispatcher is a mock of a dispatcher that can be reused for testing
type MockDispatcher struct {
	mock.Mock
}

// Start mocks Start
func (m *MockDispatcher) Start() {
	m.Called()
}

// Stop mocks Stop
func (m *MockDispatcher) Stop() {
	m.Called()
}

// Send mocks Send
func (m *MockDispatcher) Send(notification ...dispatcher.Notification) {
	m.Called(notification)
}

// Subscribers mocks Subscribers
func (m *MockDispatcher) Subscribers() []dispatcher.Subscriber {
	args := m.Called()
	return args.Get(0).([]dispatcher.Subscriber)
}

// Register mocks Register
func (m *MockDispatcher) Register(subscriber dispatcher.Subscriber) {
	m.Called(subscriber)
}

// Close mocks Close
func (m *MockDispatcher) Close(subscriber dispatcher.Subscriber) {
	m.Called(subscriber)
}

type MockTransport struct {
	ResponseStatusCode int
	ResponseBody       string
	ShouldReturnError  bool
}

func MockHTTPClientWithResponseCode(responseCode int) *http.Client {
	client := http.DefaultClient
	client.Transport = &MockTransport{
		ResponseStatusCode: responseCode,
	}
	return client
}

func DefaultMockHTTPClient() *http.Client {
	client := http.DefaultClient
	client.Transport = &MockTransport{}
	return client
}

func ErroringMockHTTPClient() *http.Client {
	client := http.DefaultClient
	client.Transport = &MockTransport{
		ShouldReturnError: true,
	}
	return client
}
func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	response := &http.Response{
		Header:     make(http.Header),
		Request:    req,
		StatusCode: t.ResponseStatusCode,
	}

	response.Header.Set("Content-Type", "application/json")
	response.Body = ioutil.NopCloser(strings.NewReader(t.ResponseBody))

	if t.ShouldReturnError {
		return nil, errors.New("Client error")
	}
	return response, nil
}
