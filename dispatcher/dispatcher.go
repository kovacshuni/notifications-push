package dispatcher

import (
	"reflect"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

const heartbeatMsg = "[]"
const heartbeatPeriod = 30 * time.Second

// Dispatcher forwards a new notification onto subscribers.
type Dispatcher interface {
	Start()
	Stop()
	Send(notification ...Notification)
	Subscribers() []Subscriber
	Registrator
}

// Registrator :smirk:
type Registrator interface {
	Register(subscriber Subscriber)
	Close(subscriber Subscriber)
}

// NewDispatcher creates and returns a new dispatcher
func NewDispatcher(delay time.Duration, history History) Dispatcher {
	return &dispatcher{
		delay:       delay,
		inbound:     make(chan Notification),
		subscribers: map[Subscriber]bool{},
		lock:        &sync.RWMutex{},
		history:     history,
		stopChan:    make(chan bool),
	}
}

type dispatcher struct {
	delay       time.Duration
	inbound     chan Notification
	subscribers map[Subscriber]bool
	lock        *sync.RWMutex
	history     History
	stopChan    chan bool
}

func (d *dispatcher) Start() {
	heartbeat := time.NewTimer(heartbeatPeriod)

	for {
		select {
		case notification := <-d.inbound:
			d.forwardToSubscribers(notification)
		case <-heartbeat.C:
			d.heartbeat()
		case <-d.stopChan:
			heartbeat.Stop()
			return
		}

		heartbeat.Reset(heartbeatPeriod)
	}
}

func (d *dispatcher) forwardToSubscribers(notification Notification) {
	log.WithField("tid", notification.PublishReference).WithField("resource", notification.APIURL).Info("Forwarding to subscribers.")

	d.lock.RLock()
	defer d.lock.RUnlock()

	for sub := range d.subscribers {
		sub.send(notification)
	}
	d.history.Push(notification)
}

func (d *dispatcher) heartbeat() {
	d.lock.RLock()
	defer d.lock.RUnlock()
	for sub := range d.subscribers {
		sub.NotificationChannel() <- heartbeatMsg
	}
}

func (d *dispatcher) Stop() {
	d.stopChan <- true
}

func (d *dispatcher) Send(notifications ...Notification) {
	go func() {
		d.delayForCache()
		for _, n := range notifications {
			d.inbound <- n
		}
	}()
}

func (d *dispatcher) delayForCache() {
	time.Sleep(d.delay)
}

func (d *dispatcher) Register(subscriber Subscriber) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.subscribers[subscriber] = true
	log.WithField("subscriber", subscriber.Address()).WithField("subscriberType", reflect.TypeOf(subscriber).Elem().Name()).Info("Registered new subscriber")
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
	log.WithField("subscriber", subscriber.Address()).WithField("subscriberType", reflect.TypeOf(subscriber).Elem().Name()).Info("Unregistered subscriber")
}
