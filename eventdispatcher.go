package main

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"time"

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
	ApiUrl string `json:"apiUrl"`
	Id     string `json:"id"`
	Type   string `json:"type"`
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
		Id:     "http://www.ft.com/thing/" + cmsPubEvent.UUID,
		ApiUrl: "http://api.ft.com/content/" + cmsPubEvent.UUID,
	}
}

const heartbeatMsg = "[]"
const heartbeatPeriod = 60

func newTimer() *time.Timer {
	return time.NewTimer(heartbeatPeriod * time.Second)
}

func resetTimer(timer *time.Timer) *time.Timer {
	if ok := timer.Reset(heartbeatPeriod * time.Second); !ok {
		infoLogger.Println("Resetting timer failed. Creating new timer")
		timer = newTimer()
	}
	return timer
}

func (d EventDispatcher) distributeEvents() {
	heartbeat := newTimer()
	for {
		select {
		case msg := <-d.incoming:
			for sub, _ := range d.subscribers {
				select {
				case sub <- msg:
				default:
					//TODO monitor this
					infoLogger.Printf("listener too far behind - message dropped")
				}
			}
			resetTimer(heartbeat)
		case <-heartbeat.C:
			infoLogger.Println("Heartbeat fired")
			for sub, _ := range d.subscribers {
				sub <- heartbeatMsg
			}
			heartbeat = newTimer()
		case subscriber := <-d.addSubscriber:
			log.Printf("New subscriber")
			d.subscribers[subscriber] = Subscriber{}
		case subscriber := <-d.removeSubscriber:
			delete(d.subscribers, subscriber)
			log.Printf("Subscriber left")
		}
	}
}
