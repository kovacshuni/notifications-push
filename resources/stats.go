package resources

import (
	"encoding/json"
	"net/http"

	"github.com/Financial-Times/notifications-push/dispatcher"
	log "github.com/Financial-Times/go-logger"
)

type subscriptionStats struct {
	NrOfSubscribers int                     `json:"nrOfSubscribers"`
	Subscribers     []dispatcher.Subscriber `json:"subscribers"`
}

// Stats returns subscriber stats
func Stats(dispatcher dispatcher.Dispatcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		subscribers := dispatcher.Subscribers()

		stats := subscriptionStats{
			NrOfSubscribers: len(subscribers),
			Subscribers:     subscribers,
		}

		bytes, err := json.Marshal(stats)
		if err != nil {
			log.WithError(err).Warn("Error in marshalling stats information", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-type", "application/json")
		b, err := w.Write(bytes)
		if b == 0 {
			log.Warn("Response written to HTTP was empty.")
		}

		if err != nil {
			log.Warnf("Error writing stats to HTTP response: %v", err.Error())
		}
	}
}
