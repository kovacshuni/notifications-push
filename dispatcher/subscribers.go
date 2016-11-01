package dispatcher

import "time"

type Subscriber struct {
	NotificationChannel chan string
	Addr                string
	Since               time.Time
	IsMonitor           bool
}

func (d *dispatcher) Register(subscriber Subscriber) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.subscribers[subscriber] = true
}

func (d *dispatcher) GetSubscribers() []Subscriber {
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
