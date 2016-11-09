package resources

import (
	"bufio"
	"net/http"
	"strconv"
	"strings"

	"github.com/Financial-Times/notifications-push/dispatcher"
	log "github.com/Sirupsen/logrus"
)

// Push handler for push subscribers
func Push(registrator dispatcher.Registrar) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

		var s dispatcher.Subscriber

		if isMonitor {
			s = dispatcher.NewMonitorSubscriber(getClientAddr(r))
		} else {
			s = dispatcher.NewStandardSubscriber(getClientAddr(r))
		}

		registrator.Register(s)
		defer registrator.Close(s)

		for {
			select {
			case notification := <-s.NotificationChannel():
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
			case <-cn.CloseNotify():
				return
			}
		}
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
