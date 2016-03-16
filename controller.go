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
	defer func() {
		c.dispatcher.removeSubscriber <- events
	}()
	for {
		_, err := bw.WriteString(<-events)
		if err != nil {
			infoLogger.Printf("[%v]", err)
			return
		}
		err = bw.WriteByte('\n')
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
