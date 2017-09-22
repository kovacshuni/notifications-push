package resources

import (
	"net/http"

	"github.com/Financial-Times/notifications-push/dispatcher"
	log "github.com/Sirupsen/logrus"
)

// History returns history data
func History(history dispatcher.History) func(w http.ResponseWriter, r *http.Request) {
	errMsg := "Serving /__history request"
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json; charset=UTF-8")

		historyJSON, err := dispatcher.MarshalNotificationsJSON(history.Notifications())
		if err != nil {
			log.WithError(err).Warn(errMsg)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		_, err = w.Write(historyJSON)
		if err != nil {
			log.WithError(err).Warn(errMsg)
			http.Error(w, "", http.StatusInternalServerError)
		}
	}
}
