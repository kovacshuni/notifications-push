package main

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/Financial-Times/message-queue-gonsumer/consumer"
)

type eventDispatcher struct {
	incoming         chan string
	subscribers      map[chan string]subscriber
	addSubscriber    chan chan string
	removeSubscriber chan chan string
}

func newEvents() *eventDispatcher {
	incoming := make(chan string)
	subscribers := make(map[chan string]subscriber)
	addSubscriber := make(chan chan string)
	removeSubscriber := make(chan chan string)
	return &eventDispatcher{incoming, subscribers, addSubscriber, removeSubscriber}
}

type notification struct {
	APIURL string `json:"apiUrl"`
	ID     string `json:"id"`
	Type   string `json:"type"`
}

type subscriber struct{}

var whitelist = regexp.MustCompile("^http://(methode-article|wordpress-article)-transformer-(pr|iw)-uk-.*\\.svc\\.ft\\.com(:\\d{2,5})?/(content)/[\\w-]+.*$")

func (d eventDispatcher) receiveEvents(msg consumer.Message) {
	if strings.HasPrefix(msg.Headers["X-Request-Id"], "SYNTH") {
		return
	}
	var cmsPubEvent cmsPublicationEvent
	err := json.Unmarshal([]byte(msg.Body), &cmsPubEvent)
	if err != nil {
		warnLogger.Printf("Skipping cmsPublicationEvent [%v]: [%v].", msg.Body, err)
		return
	}
	if !whitelist.MatchString(cmsPubEvent.ContentURI) {
		infoLogger.Printf("Skipping msg with contentUri [%v]", cmsPubEvent.ContentURI)
		return
	}

	//infoLogger.Printf("CmsPublicationEvent with tid [%v] is valid.", msg.Headers["X-Request-Id"])
	n := buildNotification(cmsPubEvent)
	if n == nil {
		warnLogger.Printf("Skipping. Cannot build notification for msg: [%#v]", cmsPubEvent)
		return
	}
	bytes, err := json.Marshal([]*notification{n})
	if err != nil {
		warnLogger.Printf("Skipping notification [%#v]: [%v]", n, err)
		return
	}

	//wait 30sec for the content to be ingested before notifying the clients
	time.Sleep(30 * time.Second)

	d.incoming <- string(bytes[:])
}

func buildNotification(cmsPubEvent cmsPublicationEvent) *notification {
	if cmsPubEvent.UUID == "" {
		return nil
	}

	eventType := "UPDATE"
	if len(cmsPubEvent.Payload) == 0 {
		eventType = "DELETE"
	}
	return &notification{
		Type:   "http://www.ft.com/thing/ThingChangeType/" + eventType,
		ID:     "http://www.ft.com/thing/" + cmsPubEvent.UUID,
		APIURL: "http://api.ft.com/content/" + cmsPubEvent.UUID,
	}
}

func (d eventDispatcher) distributeEvents() {
	heartbeat := time.NewTimer(heartbeatPeriod * time.Second)
	for {
		select {
		case msg := <-d.incoming:
			for sub := range d.subscribers {
				go func(sub chan string) { sub <- msg }(sub)
			}
			resetTimer(heartbeat)
		case <-heartbeat.C:
			for sub := range d.subscribers {
				go func(sub chan string) { sub <- heartbeatMsg }(sub)
			}
			resetTimer(heartbeat)
		case s := <-d.addSubscriber:
			log.Printf("New subscriber")
			d.subscribers[s] = subscriber{}
		case s := <-d.removeSubscriber:
			delete(d.subscribers, s)
			log.Printf("Subscriber left")
		}
	}
}

const heartbeatMsg = "[]"
const heartbeatPeriod = 60

func resetTimer(timer *time.Timer) {
	timer.Reset(heartbeatPeriod * time.Second)
}
