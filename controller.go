package main

import (
	"bufio"
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
	subscriberEvent := subscriberEvent{
		ch: events,
		subscriber: subscriber{
			addr:  getClientAddr(r),
			since: time.Now(),
		},
	}
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

func getClientAddr(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		addr := strings.Split(xForwardedFor, ",")
		return addr[0] + ":" + r.Header.Get("X-Forwarded-Port")
	}
	return r.RemoteAddr
}
