package main

import (
	"bufio"
	"net/http"
)

type Controller struct {
	dispatcher *EventDispatcher
}

func (c Controller) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	bw := bufio.NewWriter(w)

	events := make(chan string)

	c.dispatcher.addSubscriber <- events
	defer func() { c.dispatcher.removeSubscriber <- events }()

	for {
		bw.WriteString(<-events)
		bw.WriteByte('\n')
		bw.Flush()
		flusher := w.(http.Flusher)
		flusher.Flush()
	}
}
