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

type Subscriber struct{}

var whitelist = regexp.MustCompile("^http://(methode-article|methode-list|wordpress-article)-transformer-(pr|iw)-uk-.*\\.svc\\.ft\\.com(:\\d{2,5})?/(content)/[\\w-]+.*$")

func (d EventDispatcher) receiveEvents(msg consumer.Message) {
	if strings.HasPrefix(msg.Headers["X-Request-Id"], "SYNTH") {
		return
	}
	var jsonMsg map[string]interface{}
	err := json.Unmarshal([]byte(msg.Body), &jsonMsg)
	if err != nil {
		warnLogger.Printf("Skipping: [%v]", err)
		return
	}

	contentUri, ok := jsonMsg["contentUri"].(string)
	if !ok || !whitelist.MatchString(contentUri) {
		infoLogger.Printf("Skipping msg with contentUri [%v]", jsonMsg["contentUri"])
		return
	}

	//TODO use []byte instead of string
	notification := buildNotification(jsonMsg)
	if notification == "" {
		warnLogger.Printf("Cannot build notification for msg: [%v]", jsonMsg)
		return
	}
	d.incoming <- notification

}

//TODO make this complete
func buildNotification(jsonMsg map[string]interface{}) string {
	relativeUrl, ok := jsonMsg["relativeUrl"].(string)
	if !ok {
		return ""
	}
	apiUrl := "http://api.ft.com" + relativeUrl
	notification := make(map[string]string)
	notification["apiUrl"] = apiUrl
	result, err := json.Marshal(notification)
	if err != nil {
		warnLogger.Printf("Marshalling failed: [%v]", err)
		return ""
	}
	return string(result[:])
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
