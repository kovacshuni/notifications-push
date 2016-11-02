package resources

import (
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/Financial-Times/notifications-push/dispatcher"
	log "github.com/Sirupsen/logrus"
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
}

func (s *externalSubscriber) MarshalJSON() ([]byte, error) {
	return json.Marshal(newSubscriberPayload(s))
}

func (m *monitorSubscriber) MarshalJSON() ([]byte, error) {
	return json.Marshal(newSubscriberPayload(m))
}

type SubscriberPayload struct {
	Address            string `json:"address"`
	Since              string `json:"since"`
	ConnectionDuration string `json:"connectionDuration"`
	Type               string `json:"type"`
}

func newSubscriberPayload(s Subscriber) *SubscriberPayload {
	return &SubscriberPayload{
		Address:            s.address(),
		Since:              s.since().Format(time.StampMilli),
		ConnectionDuration: time.Since(s.since()).String(),
		Type:               reflect.TypeOf(s).Elem().String(),
	}
}
