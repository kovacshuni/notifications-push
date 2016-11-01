package dispatcher

import (
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
		sub.NotificationChannel() <- heartbeatMsg
	}
}

func (d *dispatcher) delayForCache(notification Notification) {
	time.Sleep(time.Duration(d.delay) * time.Second)
}

func (d *dispatcher) forwardToSubscribers(notification Notification) {
	log.WithField("tid", notification.PublishReference).WithField("id", notification.ID).Info("Forwarding to subscribers.")

	d.lock.RLock()
	defer d.lock.RUnlock()

	for sub := range d.subscribers {
		sub.send(notification)
	}
}

func (d *dispatcher) Register(subscriber Subscriber) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.subscribers[subscriber] = true
}

func (d *dispatcher) Subscribers() []Subscriber {
	d.lock.RLock()
	defer d.lock.RUnlock()

	var subs []Subscriber
	for sub := range d.subscribers {
		subs = append(subs, sub)
	}
	return subs
}

func (d *dispatcher) Close(subscriber Subscriber) {
	d.lock.Lock()
	defer d.lock.Unlock()

	delete(d.subscribers, subscriber)
}
