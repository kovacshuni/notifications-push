package dispatcher

import (
	"encoding/json"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

type dispatcher struct {
	delay       int
	inbound     chan Notification
	subscribers map[Subscriber]bool
	lock        *sync.RWMutex
	history     History
}

func (d *dispatcher) Send(notification Notification) {
	log.WithField("tid", notification.PublishReference).WithField("id", notification.ID).Infof("Received event. Waiting configured delay (%vs).", d.delay)
	go func() {
		d.delayForCache(notification)
		d.inbound <- notification
	}()
}

func (d *dispatcher) heartbeat() {
	d.lock.RLock()
	defer d.lock.RUnlock()

	for sub := range d.subscribers {
		sub.NotificationChannel <- heartbeatMsg
	}
}

func (d *dispatcher) delayForCache(notification Notification) {
	time.Sleep(time.Duration(d.delay) * time.Second)
}

func (d *dispatcher) forwardToSubscribers(notification Notification) {
	log.WithField("tid", notification.PublishReference).WithField("id", notification.ID).Info("Forwarding to subscribers.")

	d.lock.RLock()
	defer d.lock.RUnlock()

	regular, monitor, err := marshal(notification)
	if err != nil {
		log.WithError(err).Error("Failed to marshal notification!")
		return
	}

	for sub := range d.subscribers {
		if sub.IsMonitor {
			sub.NotificationChannel <- monitor
			continue
		}

		sub.NotificationChannel <- regular
	}
}

func marshal(notification Notification) (string, string, error) {
	regular, err := json.Marshal(notification)

	notification.PublishReference = ""
	notification.LastModified = ""
	monitor, _ := json.Marshal(notification)

	if err != nil {
		return "", "", err
	}

	return string(regular), string(monitor), err
}
