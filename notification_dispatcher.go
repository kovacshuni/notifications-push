package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

const heartbeatMsg = "[]"
const heartbeatPeriod = 30 * time.Second

type notificationDispatcher struct {
	incomingNotifications chan notification
	subscribers           map[subscriber]bool
	delay                 int
	notificationHistory   *notificationHistory
	subscriberMutex       *sync.RWMutex
}

type subscriber struct {
	notificationChannel chan string
	Addr                string
	Since               time.Time
	isMonitor           bool
}

func newNotificationDispatcher(delay int, historySize int) *notificationDispatcher {
	incomingNotifications := make(chan notification)
	subscribers := make(map[subscriber]bool)
	history := newNotificationHistory(historySize)
	return &notificationDispatcher{incomingNotifications, subscribers, delay, history, &sync.RWMutex{}}
}

func (d notificationDispatcher) start() {
	heartbeat := time.NewTimer(heartbeatPeriod)
	for {
		select {
		case n := <-d.incomingNotifications:
			d.sendNotification(n)
		case <-heartbeat.C:
			d.sendHeartbeat()
		}
		heartbeat.Reset(heartbeatPeriod)
	}
}

func (d *notificationDispatcher) sendNotification(n notification) {

	defaultMsg, err := d.buildDefaultMessage(n)
	if err != nil {
		log.Warnf("Skipping notification: Notification [%#v]: [%v]", n.PublishReference, n, err)
		return
	}

	monitorMsg, err := d.buildMonitorMessage(n)
	if err != nil {
		log.Warnf("Skipping notification: Notification [%#v]: [%v]", n.PublishReference, n, err)
		return
	}

	d.subscriberMutex.RLock()
	defer d.subscriberMutex.RUnlock()

	for sub, _ := range d.subscribers {
		if sub.isMonitor {
			sub.send(monitorMsg)
		} else {
			sub.send(defaultMsg)
		}
	}

	d.notificationHistory.add(n)
}

func (d notificationDispatcher) buildDefaultMessage(n notification) (string, error) {
	n.PublishReference = ""
	n.LastModified = ""
	return d.buildMessage(n)
}

func (d notificationDispatcher) buildMonitorMessage(n notification) (string, error) {
	return d.buildMessage(n)
}

func (d notificationDispatcher) buildMessage(n notification) (string, error) {
	json, err := json.Marshal([]notification{n})
	if err != nil {
		return "", err
	}
	return string(json[:]), nil
}

func (d *notificationDispatcher) sendHeartbeat() {
	for sub, _ := range d.subscribers {
		sub.send(heartbeatMsg)
	}
}

func (d notificationDispatcher) dispatch(n notification) {
	time.Sleep(time.Duration(d.delay) * time.Second)
	log.Infof("Notifying clients about tid=[%v] id=[%v].", n.PublishReference, n.APIURL)
	d.incomingNotifications <- n
}

func (d *notificationDispatcher) add(s subscriber) {
	d.subscriberMutex.Lock()
	defer d.subscriberMutex.Unlock()

	d.subscribers[s] = true
	log.Infof("New subscriber [%s].", s.Addr)
}

func (d *notificationDispatcher) remove(s subscriber) {
	d.subscriberMutex.Lock()
	defer d.subscriberMutex.Unlock()

	delete(d.subscribers, s)
	log.Infof("Subscriber left [%s].", s)
}

func (s *subscriber) send(msg string) {
	select {
	case s.notificationChannel <- msg:
	default:
		log.Warnf("Subscriber [%v] lagging behind.", s)
	}
}

func (s subscriber) String() string {
	return fmt.Sprintf("Addr=[%s]. Since=[%s]. Connection duration=[%s].", s.Addr, s.Since.Format(time.StampMilli), time.Since(s.Since))
}

func (s subscriber) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{ "addr" : "%s", "since" : "%s", "connectionDuration": "%s", "isMonitor":"%v" }`, s.Addr, s.Since.Format(time.StampMilli), time.Since(s.Since), s.isMonitor)), nil
}
