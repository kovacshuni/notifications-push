package dispatcher

import (
	"encoding/json"
	"reflect"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Subscriber represents the interface of a generic subscriber to a push stream
type Subscriber interface {
	send(n Notification) error
	NotificationChannel() chan string
	writeOnMsgChannel(string)
	Address() string
	Since() time.Time
}

// StandardSubscriber implements a standard subscriber
type StandardSubscriber struct {
	notificationChannel chan string
	addr                string
	sinceTime           time.Time
}

// NewStandardSubscriber returns a new instance of a standard subscriber
func NewStandardSubscriber(address string) *StandardSubscriber {
	notificationChannel := make(chan string, 16)
	return &StandardSubscriber{
		notificationChannel: notificationChannel,
		addr:                address,
		sinceTime:           time.Now(),
	}
}

// Address returns the IP address of the standard subscriber
func (s *StandardSubscriber) Address() string {
	return s.addr
}

// Since returns the time since a subscriber have been registered
func (s *StandardSubscriber) Since() time.Time {
	return s.sinceTime
}

func (s *StandardSubscriber) send(n Notification) error {
	notificationMsg, err := buildStandardNotificationMsg(n)
	if err != nil {
		return err
	}
	s.writeOnMsgChannel(notificationMsg)
	return nil
}

// NotificationChannel returns the channel that can be used to send
// notifications to the standard subscriber
func (s *StandardSubscriber) NotificationChannel() chan string {
	return s.notificationChannel
}

func (s *StandardSubscriber) writeOnMsgChannel(msg string) {
	select {
	case s.notificationChannel <- msg:
	default:
		log.WithField("subscriber", s.Address()).WithField("message", msg).Warn("Subscriber lagging behind...")
	}
}

func buildStandardNotificationMsg(n Notification) (string, error) {
	n.PublishReference = ""
	n.LastModified = ""

	return buildNotificationMsg(n)
}

func buildNotificationMsg(n Notification) (string, error) {
	jsonNotification, err := json.Marshal([]Notification{n})

	if err != nil {
		return "", err
	}

	return string(jsonNotification), err
}

// MonitorSubscriber implements a Monitor subscriber
type MonitorSubscriber struct {
	*StandardSubscriber
}

// NewMonitorSubscriber returns a new instance of a Monitor subscriber
func NewMonitorSubscriber(address string) *MonitorSubscriber {
	return &MonitorSubscriber{NewStandardSubscriber(address)}
}

func (m *MonitorSubscriber) send(n Notification) error {
	notificationMsg, err := buildMonitorNotificationMsg(n)
	if err != nil {
		return err
	}
	m.writeOnMsgChannel(notificationMsg)
	return nil
}

func buildMonitorNotificationMsg(n Notification) (string, error) {
	return buildNotificationMsg(n)
}

// MarshalJSON returns the JSON representation of a StandardSubscriber
func (s *StandardSubscriber) MarshalJSON() ([]byte, error) {
	return json.Marshal(newSubscriberPayload(s))
}

// MarshalJSON returns the JSON representation of a MonitorSubscriber
func (m *MonitorSubscriber) MarshalJSON() ([]byte, error) {
	return json.Marshal(newSubscriberPayload(m))
}

// SubscriberPayload is the JSON representation of a generic subscriber
type SubscriberPayload struct {
	Address            string `json:"address"`
	Since              string `json:"since"`
	ConnectionDuration string `json:"connectionDuration"`
	Type               string `json:"type"`
}

func newSubscriberPayload(s Subscriber) *SubscriberPayload {
	return &SubscriberPayload{
		Address:            s.Address(),
		Since:              s.Since().Format(time.StampMilli),
		ConnectionDuration: time.Since(s.Since()).String(),
		Type:               reflect.TypeOf(s).Elem().String(),
	}
}
