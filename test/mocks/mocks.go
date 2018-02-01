package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/Financial-Times/notifications-push/dispatch"
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
func (m *MockDispatcher) Send(notification ...dispatch.Notification) {
	m.Called(notification)
}

// Subscribers mocks Subscribers
func (m *MockDispatcher) Subscribers() []dispatch.Subscriber {
	args := m.Called()
	return args.Get(0).([]dispatch.Subscriber)
}

// Register mocks Register
func (m *MockDispatcher) Register(subscriber dispatch.Subscriber) {
	m.Called(subscriber)
}

// Close mocks Close
func (m *MockDispatcher) Close(subscriber dispatch.Subscriber) {
	m.Called(subscriber)
}
