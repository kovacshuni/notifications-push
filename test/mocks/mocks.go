package mocks

import (
	"github.com/Financial-Times/notifications-push/dispatcher"
	"github.com/stretchr/testify/mock"
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
