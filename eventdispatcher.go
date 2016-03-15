package main

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"

	"github.com/Financial-Times/message-queue-gonsumer/consumer"
)

type EventDispatcher struct {
	incoming         chan string
	subscribers      map[chan string]Subscriber
	addSubscriber    chan chan string
	removeSubscriber chan chan string
}

func NewEvents() *EventDispatcher {
	incoming := make(chan string)
	subscribers := make(map[chan string]Subscriber)
	addSubscriber := make(chan chan string)
	removeSubscriber := make(chan chan string)
	return &EventDispatcher{incoming, subscribers, addSubscriber, removeSubscriber}
}

type notification struct {
	ApiUrl string
	Id     string
	Type   string
}

type Subscriber struct{}

var whitelist = regexp.MustCompile("^http://(methode-article|wordpress-article)-transformer-(pr|iw)-uk-.*\\.svc\\.ft\\.com(:\\d{2,5})?/(content)/[\\w-]+.*$")

func (d EventDispatcher) receiveEvents(msg consumer.Message) {
	if strings.HasPrefix(msg.Headers["X-Request-Id"], "SYNTH") {
		return
	}
	var cmsPubEvent cmsPublicationEvent
	err := json.Unmarshal([]byte(msg.Body), &cmsPubEvent)
	if err != nil {
		warnLogger.Printf("Skipping cmsPublicationEvent [%v]: [%v].", msg.Body, err)
		return
	}
	if !whitelist.MatchString(cmsPubEvent.ContentUri) {
		infoLogger.Printf("Skipping msg with contentUri [%v]", cmsPubEvent.ContentUri)
		return
	}

	notification := buildNotification(cmsPubEvent)
	if notification == nil {
		warnLogger.Printf("Skipping. Cannot build notification for msg: [%#v]", cmsPubEvent)
		return
	}
	bytes, err := json.Marshal(notification)
	if err != nil {
		warnLogger.Printf("Skipping notification [%#v]: [%v]", notification, err)
		return
	}
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
		ApiUrl: "http://www.ft.com/thing/ThingChangeType/" + eventType,
		Id:     "http://api.ft.com/content/" + cmsPubEvent.UUID,
		Type:   "http://www.ft.com/thing/ThingChangeType/" + eventType,
	}
}

func (d EventDispatcher) distributeEvents() {
	for {
		select {
		case msg := <-d.incoming:
			for sub, _ := range d.subscribers {
				select {
				case sub <- msg:
				default:
					infoLogger.Printf("listener too far behind - message dropped")
				}
			}
		case subscriber := <-d.addSubscriber:
			log.Printf("New subscriber")
			d.subscribers[subscriber] = Subscriber{}
		case subscriber := <-d.removeSubscriber:
			delete(d.subscribers, subscriber)
			log.Printf("Subscriber left")
		}
	}
}
