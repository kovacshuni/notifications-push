package main

import (
	"encoding/json"
	"log"
	"time"
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

func (d EventDispatcher) receiveEvents() {
	// this simulates some incoming events that we wish to distribute to consumers
	go func() {
		for {
			str, err := json.Marshal(struct {
				Date time.Time `json:"date"`
			}{time.Now()})
			if err != nil {
				panic(err)
			}
			d.incoming <- string(str)
			time.Sleep(2 * time.Second)
		}
	}()
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
