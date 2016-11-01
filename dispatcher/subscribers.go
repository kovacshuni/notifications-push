package dispatcher

import (
	"encoding/json"
	"reflect"
	"time"
)

type Subscriber interface {
	send(n Notification) error
	NotificationChannel() chan string
	address() string
	since() time.Time
}

type externalSubscriber struct {
	notificationChannel chan string
	addr                string
	sinceTime           time.Time
}

func NewExternalSubscriber(address string) *externalSubscriber {
	notificationChannel := make(chan string, 16)
	return &externalSubscriber{
		notificationChannel: notificationChannel,
		addr:                address,
		sinceTime:           time.Now(),
	}
}

func (s *externalSubscriber) address() string {
	return s.addr
}

func (s *externalSubscriber) since() time.Time {
	return s.sinceTime
}

func (s *externalSubscriber) send(n Notification) error {
	notificationMsg, err := s.marshal(n)
	if err != nil {
		return err
	}
	s.notificationChannel <- notificationMsg
	return nil
}

func (s *externalSubscriber) marshal(n Notification) (string, error) {
	n.PublishReference = ""
	n.LastModified = ""

	jsonNotification, err := json.Marshal(n)

	if err != nil {
		return "", err
	}

	return string(jsonNotification), err
}

func (s *externalSubscriber) NotificationChannel() chan string {
	return s.notificationChannel
}

func (s *externalSubscriber) MarshalJSON() ([]byte, error) {
	return json.Marshal(newSubscriberPayload(s))
}

type SubscriberPayload struct {
	Address            string `json:"address"`
	Since              string `json:"since"`
	ConnectionDuration string `json:"connectionDuration"`
	Type               string `json:"type"`
}

func newSubscriberPayload(s Subscriber) *SubscriberPayload {
	return &SubscriberPayload{
		Address:            s.address(),
		Since:              s.since().Format(time.StampMilli),
		ConnectionDuration: time.Since(s.since()).String(),
		Type:               reflect.TypeOf(s).Elem().Name(),
	}
}
