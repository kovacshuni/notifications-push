package dispatcher

import (
	"encoding/json"
	"time"
)

type Subscriber interface {
	send(n Notification) error
	NotificationChannel() chan string
	address() string
	since() time.Time
}

// Standard Subscriber implementation
type standardSubscriber struct {
	notificationChannel chan string
	addr                string
	sinceTime           time.Time
}

func NewStandardSubscriber(address string) *standardSubscriber {
	notificationChannel := make(chan string, 16)
	return &standardSubscriber{
		notificationChannel: notificationChannel,
		addr:                address,
		sinceTime:           time.Now(),
	}
}

func (s *standardSubscriber) address() string {
	return s.addr
}

func (s *standardSubscriber) since() time.Time {
	return s.sinceTime
}

func (s *standardSubscriber) send(n Notification) error {
	notificationMsg, err := buildStandardNotificationMsg(n)
	if err != nil {
		return err
	}
	s.notificationChannel <- notificationMsg
	return nil
}

func buildStandardNotificationMsg(n Notification) (string, error) {
	n.PublishReference = ""
	n.LastModified = ""

	return buildNotificationMsg(n)
}

func buildNotificationMsg(n Notification) (string, error) {
	jsonNotification, err := json.Marshal(n)

	if err != nil {
		return "", err
	}

	return string(jsonNotification), err
}

func (s *standardSubscriber) NotificationChannel() chan string {
	return s.notificationChannel
}

// Monitor Subscriber implementation
type monitorSubscriber struct {
	*standardSubscriber
}

func NewMonitorSubscriber(address string) *monitorSubscriber {
	return &monitorSubscriber{NewStandardSubscriber(address)}
}

func (m *monitorSubscriber) send(n Notification) error {
	notificationMsg, err := buildMonitorNotificationMsg(n)
	if err != nil {
		return err
	}
	m.notificationChannel <- notificationMsg
	return nil
}

func buildMonitorNotificationMsg(n Notification) (string, error) {
	return buildNotificationMsg(n)
}
