package main

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

const evPrefix = "data: "
const errMsgPrefix = "Serving /notifications request: [%v]"

type handler struct {
	resource           string
	dispatcher         *eventDispatcher
	notificationsCache *uniqueue
	apiBaseURL         string
	internalBaseURL    string
}

func newHandler(resource string, dispatcher *eventDispatcher, notificationsCache *uniqueue, apiBaseURL string) handler {
	return handler{resource, dispatcher, notificationsCache, apiBaseURL, apiBaseURL + "/__notifications-push"}
}

type stats struct {
	NrOfSubscribers int          `json:"nrOfSubscribers"`
	Subscribers     []subscriber `json:"subscribers"`
}

func (h handler) notificationsPush(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/event-stream")
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
			Addr:  getClientAddr(r),
			Since: time.Now(),
		},
	}
	h.dispatcher.addSubscriber <- subscriberEvent
	defer func() {
		h.dispatcher.removeSubscriber <- subscriberEvent
	}()

	for {
		select {
		case <-cn.CloseNotify():
			return
		case event := <-events:
			_, err := bw.WriteString(evPrefix + event + "\n\n")
			if err != nil {
				log.Infof("[%v]", err)
				return
			}
			err = bw.Flush()
			if err != nil {
				log.Infof("[%v]", err)
				return
			}
			flusher := w.(http.Flusher)
			flusher.Flush()
		}
	}
}

func (h handler) notifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	var pageUpp notificationsPageUpp
	isEmpty := r.URL.Query().Get("empty")

	if isEmpty == "true" {
		pageUpp = newNotificationsPageUpp([]notificationUPP{}, r.URL.RequestURI(), h.apiBaseURL, h.resource, h.internalBaseURL)
	} else {
		it := h.notificationsCache.items()
		ns := make([]notificationUPP, len(it))
		for i := range it {
			ns[i] = *it[i]
		}
		pageUpp = newNotificationsPageUpp(ns, r.URL.RequestURI(), h.apiBaseURL, h.resource, h.internalBaseURL)
	}
	bytes, err := json.Marshal(pageUpp)
	if err != nil {
		log.Warnf(errMsgPrefix, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		log.Warnf(errMsgPrefix, err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (h handler) stats(w http.ResponseWriter, r *http.Request) {
	subscribers := []subscriber{}
	for _, s := range h.dispatcher.subscribers {
		subscribers = append(subscribers, s)
	}
	stats := stats{
		NrOfSubscribers: len(h.dispatcher.subscribers),
		Subscribers:     subscribers,
	}
	bytes, err := json.Marshal(stats)
	if err != nil {
		log.Warnf("[%v]", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	b, err := w.Write(bytes)
	if b == 0 {
		log.Warnf("Response written to HTTP was empty.")
	}
	if err != nil {
		log.Warnf("Error writing stats to HTTP response: %v", err.Error())
	}
}

func getClientAddr(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		addr := strings.Split(xForwardedFor, ",")
		return addr[0]
	}
	return r.RemoteAddr
}
