package main

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type controller struct {
	dispatcher *eventDispatcher
}

func (c controller) notifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	cn, ok := w.(http.CloseNotifier)
	if !ok {
		http.Error(w, "Cannot stream.", http.StatusInternalServerError)
		return
	}

	bw := bufio.NewWriter(w)

	events := make(chan string, 16)
	subscriberEvent := subscriberEvent{ch: events, subscriber: buildSubscriber(r.Header)}
	c.dispatcher.addSubscriber <- subscriberEvent
	defer func() {
		c.dispatcher.removeSubscriber <- subscriberEvent
	}()

	for {
		select {
		case <-cn.CloseNotify():
			return
		case event := <-events:
			_, err := bw.WriteString(event + "\n")
			if err != nil {
				infoLogger.Printf("[%v]", err)
				return
			}
			err = bw.Flush()
			if err != nil {
				infoLogger.Printf("[%v]", err)
				return
			}
			flusher := w.(http.Flusher)
			flusher.Flush()
		}
	}
}

func buildSubscriber(header http.Header) subscriber {
	addrs := strings.Split(header.Get("X-Forwarded-For"), ",")
	return subscriber{
		Addr:  addrs[0] + ":" + header.Get("X-Forwarded-Port"),
		Since: time.Now(),
	}
}

type stats struct {
	NrOfSubscribers int          `json:"nrOfSubscribers"`
	Subscribers     []subscriber `json:"subscribers"`
}

func (c controller) stats(w http.ResponseWriter, r *http.Request) {
	var subscribers []subscriber
	for _, s := range c.dispatcher.subscribers {
		subscribers = append(subscribers, s)
	}
	stats := stats{
		NrOfSubscribers: len(c.dispatcher.subscribers),
		Subscribers:     subscribers,
	}
	bytes, err := json.Marshal(stats)
	if err != nil {
		warnLogger.Printf("[%v]", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.Write(bytes)
}
