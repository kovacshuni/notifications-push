package main

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
)

const heartbeatMsg = "[]"
const heartbeatPeriod = 30

type eventDispatcher struct {
	incoming         chan string
	subscribers      map[chan string]subscriber
	addSubscriber    chan subscriberEvent
	removeSubscriber chan subscriberEvent
}

type subscriberEvent struct {
	ch         chan string
	subscriber subscriber
}

type subscriber struct {
	Addr  string
	Since time.Time
}

func newDispatcher() *eventDispatcher {
	incoming := make(chan string)
	subscribers := make(map[chan string]subscriber)
	addSubscriber := make(chan subscriberEvent)
	removeSubscriber := make(chan subscriberEvent)
	return &eventDispatcher{incoming, subscribers, addSubscriber, removeSubscriber}
}

func (d eventDispatcher) distributeEvents() {
	heartbeat := time.NewTimer(heartbeatPeriod * time.Second)
	for {
		select {
		case msg := <-d.incoming:
			d.sendMsg(msg)
			resetTimer(heartbeat)
		case <-heartbeat.C:
			d.sendHeartBeat()
			resetTimer(heartbeat)
		case s := <-d.addSubscriber:
			log.Infof("New subscriber [%s].", s.subscriber.Addr)
			d.subscribers[s.ch] = s.subscriber
			s.ch <- heartbeatMsg
		case s := <-d.removeSubscriber:
			delete(d.subscribers, s.ch)
			log.Infof("Subscriber left [%s].", s.subscriber)
		}
	}
}

func (d eventDispatcher) sendMsg(msg string) {
	for subCh, sub := range d.subscribers {
		select {
		case subCh <- msg:
		default:
			log.Warnf("Subscriber [%v] lagging behind.", sub)
		}
	}
}

func (d eventDispatcher) sendHeartBeat() {
	for subCh, sub := range d.subscribers {
		select {
		case subCh <- heartbeatMsg:
		default:
			log.Warnf("Subscriber [%v] lagging behind when sending heartbeat.", sub)
		}
	}
}

func resetTimer(timer *time.Timer) {
	timer.Reset(heartbeatPeriod * time.Second)
}

func (s subscriber) String() string {
	return fmt.Sprintf("Addr=[%s]. Since=[%s]. Connection duration=[%s].", s.Addr, s.Since.Format(time.StampMilli), time.Since(s.Since))
}

func (s subscriber) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{ "addr" : "%s", "since" : "%s", "connectionDuration": "%s" }`, s.Addr, s.Since.Format(time.StampMilli), time.Since(s.Since))), nil
}
