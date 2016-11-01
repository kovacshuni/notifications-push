package mocks

import (
	"github.com/Financial-Times/notifications-push/dispatcher"
	"github.com/stretchr/testify/mock"
)

type MockDispatcher struct {
	mock.Mock
}

func Any(test interface{}) bool {
	return true
}

// Dispatcher Mocks

func (m MockDispatcher) Start() {
	m.Called()
}

func (m MockDispatcher) Send(notification dispatcher.Notification) {
	m.Called(notification)
}

func (m MockDispatcher) GetSubscribers() []dispatcher.Subscriber {
	args := m.Called()
	return args.Get(0).([]dispatcher.Subscriber)
}

func (m MockDispatcher) Register(subscriber dispatcher.Subscriber) {
	m.Called(subscriber)
}

func (m MockDispatcher) Close(subscriber dispatcher.Subscriber) {
	m.Called(subscriber)
}
