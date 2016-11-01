package dispatcher

import (
	"sync"
	"time"
)

const heartbeatMsg = "[]"
const heartbeatPeriod = 30 * time.Second

// NewDispatcher creates and returns a new dispatcher
func NewDispatcher(delay int, history History) Dispatcher {
	return &dispatcher{
		delay:       delay,
		inbound:     make(chan Notification),
		subscribers: map[Subscriber]bool{},
		lock:        &sync.RWMutex{},
		history:     history,
	}
}

func (d *dispatcher) Start() {
	heartbeat := time.NewTimer(heartbeatPeriod)

	for {
		select {
		case notification := <-d.inbound:
			d.forwardToSubscribers(notification)
		case <-heartbeat.C:
			d.heartbeat()
		}

		heartbeat.Reset(heartbeatPeriod)
	}
}

// Dispatcher forwards a new notification onto subscribers.
type Dispatcher interface {
	Start()
	Send(notification Notification)
	GetSubscribers() []Subscriber
	Registrator
}

// Registrator :smirk:
type Registrator interface {
	Register(subscriber Subscriber)
	Close(subscriber Subscriber)
}
