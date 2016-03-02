package main

import (
	"log"
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

func (d EventDispatcher) receiveEvents(msg consumer.Message) {
	d.incoming <- msg.Body // temporary tostring :)
}

func (d EventDispatcher) distributeEvents() {
	for {
		select {
		case msg := <-d.incoming:
			for sub, _ := range d.subscribers {
				select {
				case sub <- msg:
				default:
					log.Printf("listener too far behind - message dropped")
				}
			}
		case subscriber := <-d.addSubscriber:
			d.subscribers[subscriber] = Subscriber{}
		case subscriber := <-d.removeSubscriber:
			delete(d.subscribers, subscriber)
		}
	}
}
