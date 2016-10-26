package main

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

const errMsgPrefix = "Serving /notifications request: [%v]"

type httpHandler struct {
	dispatcher *notificationDispatcher
}

func newHttpHandler(dispatcher *notificationDispatcher) httpHandler {
	return httpHandler{dispatcher}
}

type subscriptionStats struct {
	NrOfSubscribers int          `json:"nrOfSubscribers"`
	Subscribers     []subscriber `json:"subscribers"`
}

func (h httpHandler) notificationsPush(w http.ResponseWriter, r *http.Request) {
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

	monitorParam := r.URL.Query().Get("monitor")
	isMonitor, _ := strconv.ParseBool(monitorParam)

	notificationChannel := make(chan string, 16)
	s := subscriber{
		notificationChannel: notificationChannel,
		Addr:                getClientAddr(r),
		Since:               time.Now(),
		isMonitor:           isMonitor,
	}
	h.dispatcher.add(s)
	defer h.dispatcher.remove(s)

	for {
		select {
		case <-cn.CloseNotify():
			return
		case notification := <-notificationChannel:
			_, err := bw.WriteString("data: " + notification + "\n\n")
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

func (h httpHandler) history(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	historyJson, err := json.Marshal(h.dispatcher.notificationHistory.history())
	if err != nil {
		log.Warnf(errMsgPrefix, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(historyJson)
	if err != nil {
		log.Warnf(errMsgPrefix, err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (h httpHandler) stats(w http.ResponseWriter, r *http.Request) {
	subscribers := []subscriber{}
	for s, _ := range h.dispatcher.subscribers {
		subscribers = append(subscribers, s)
	}
	stats := subscriptionStats{
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
