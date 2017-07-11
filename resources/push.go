package resources

import (
	"bufio"
	"net/http"
	"strconv"
	"strings"

	"github.com/Financial-Times/notifications-push/dispatcher"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"io"
)

const(
	ApiKeyHeaderField = "X-Api-Key"
	ApiKeyQueryParam = "apiKey"
)

//ApiKey is provided either as a request param or as a header.
func getApiKey(r *http.Request) string {
	apiKey := r.Header.Get(ApiKeyHeaderField)
	if apiKey != "" {
		return apiKey
	}

	return r.URL.Query().Get(ApiKeyQueryParam)
}

// Push handler for push subscribers
func Push(reg dispatcher.Registrar, masheryApiKeyValidationURL string, httpClient *http.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		apiKey := getApiKey(r)
		if !validApiKey(w, apiKey, masheryApiKeyValidationURL, httpClient) {
			return
		}

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

func validApiKey(w http.ResponseWriter, providedApiKey string, masheryApiKeyValidationURL string, httpClient *http.Client) bool {
	req, err := http.NewRequest("GET", masheryApiKeyValidationURL, nil)
	if err != nil {
		log.WithField("url", masheryApiKeyValidationURL).WithError(err).Error("Invalid URL for api key validation")
		http.Error(w, "Invalid URL", http.StatusInternalServerError)
		return false
	}

	req.Header.Set(ApiKeyHeaderField, providedApiKey)

	apiKeyFirstChars := ""
	if isApiKeyFirstFourCharsLoggable(providedApiKey) {
		apiKeyFirstChars = providedApiKey[0:3]
	}
	log.WithField("url", req.URL.String()).WithField("apiKeyFirstChars", apiKeyFirstChars).Info("Calling Mashery to validate api key")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.WithField("url", req.URL.String()).WithError(err).Error("Cannot send request to Mashery")
		http.Error(w, "Request to validate api key failed", http.StatusInternalServerError)
		return false
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	respStatusCode := resp.StatusCode
	if respStatusCode == http.StatusOK {
		return true
	}

	if respStatusCode == http.StatusUnauthorized {
		log.WithField("apiKeyFirstChars", apiKeyFirstChars).WithError(err).Error("Invalid api key")
		http.Error(w, "Invalid api key", http.StatusUnauthorized)
		return false
	}

	log.WithError(err).Errorf("Received unexpected status code from Mashery: %d", respStatusCode)
	http.Error(w, "Request to validate api key returned an unexpected response", http.StatusServiceUnavailable)
	return false
}

func isApiKeyFirstFourCharsLoggable(providedApiKey string) bool {
	return len(providedApiKey) > 4
}