package resources

import (
	"bufio"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Financial-Times/notifications-push/dispatcher"
	log "github.com/Sirupsen/logrus"
)

// Push handler for push subscribers
func Push(registrator dispatcher.Registrator) func(w http.ResponseWriter, r *http.Request) {
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

		notificationChannel := make(chan string, 16)
		s := dispatcher.Subscriber{
			NotificationChannel: notificationChannel,
			Addr:                getClientAddr(r),
			Since:               time.Now(),
			IsMonitor:           isMonitor,
		}

		registrator.Register(s)
		defer registrator.Close(s)

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
}

func getClientAddr(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		addr := strings.Split(xForwardedFor, ",")
		return addr[0]
	}
	return r.RemoteAddr
}
