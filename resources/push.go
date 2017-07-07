package resources

import (
	"bufio"
	"net/http"
	"strconv"
	"strings"

	"github.com/Financial-Times/notifications-push/dispatcher"
	log "github.com/Sirupsen/logrus"
)

//ApiKey is provided either as a request param or as a header.
func getApiKey(r *http.Request) string {
	apiKey := r.Header.Get("x-api-key")
	if apiKey != "" {
		return apiKey
	}

	return r.URL.Query().Get("apiKey")
}

// Push handler for push subscribers
func Push(reg dispatcher.Registrar, masheryApiKeyValidationURL string, httpClient *http.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		log.Infof("Validating api key..")
		apiKey := getApiKey(r)
		if isValid := validateApiKey(apiKey, masheryApiKeyValidationURL, httpClient, w); !isValid {
			log.Infof("Provided api key is invalid")
			return
		}

		log.Infof("Provided api key is valid")

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

		reg.Register(s)
		defer reg.Close(s)

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

func validateApiKey(providedApiKey string, masheryApiKeyValidationURL string, httpClient *http.Client, w http.ResponseWriter) bool {
	req, err := http.NewRequest("GET", masheryApiKeyValidationURL, nil)
	req.Header.Add("x-api-key", providedApiKey)
	if err != nil {
		log.WithError(err).Warn("Cannot create request")
		http.Error(w, "Invalid api key", http.StatusInternalServerError)
		return false
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.WithError(err).Warn("Cannot send request to mashery. Request url is %s", masheryApiKeyValidationURL)
		http.Error(w, "Invalid api key", http.StatusInternalServerError)
		return false
	}

	defer resp.Body.Close()

	respStatusCode := resp.StatusCode
	if respStatusCode == http.StatusOK {
		return true
	}

	if respStatusCode == http.StatusUnauthorized {
		log.WithError(err).Warn("Invalid api key")
		http.Error(w, "Invalid api key", http.StatusUnauthorized)
		return false
	}

	log.WithError(err).Warn("Received unexpected status code from Mashery: %d", respStatusCode)
	http.Error(w, "", http.StatusServiceUnavailable)
	return false
}
