package resources

import (
	"bufio"
	"net/http"
	"strconv"
	"strings"

	"fmt"
	log "github.com/Financial-Times/go-logger"
	"github.com/Financial-Times/notifications-push/dispatch"
)

const (
	apiKeyHeaderField  = "X-Api-Key"
	apiKeyQueryParam   = "apiKey"
	defaultContentType = "Article"
)

var supportedContentTypes = []string{"Article", "ContentPackage", "All"}

//ApiKey is provided either as a request param or as a header.
func getApiKey(r *http.Request) string {
	apiKey := r.Header.Get(apiKeyHeaderField)
	if apiKey != "" {
		return apiKey
	}

	return r.URL.Query().Get(apiKeyQueryParam)
}

// Push handler for push subscribers
func Push(reg dispatch.Registrar, masheryApiKeyValidationURL string, httpClient *http.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/event-stream; charset=UTF-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		apiKey := getApiKey(r)
		if isValid, errMsg, errStatusCode := isValidApiKey(apiKey, masheryApiKeyValidationURL, httpClient); !isValid {
			http.Error(w, errMsg, errStatusCode)
			return
		}

		cn, ok := w.(http.CloseNotifier)
		if !ok {
			http.Error(w, "Cannot stream.", http.StatusInternalServerError)
			return
		}

		bw := bufio.NewWriter(w)

		contentTypeParam, err := resolveContentType(r)
		if err != nil {
			log.WithError(err).Error("Invalid content type")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		monitorParam := r.URL.Query().Get("monitor")
		isMonitor, _ := strconv.ParseBool(monitorParam)

		var s dispatch.Subscriber

		if isMonitor {
			s = dispatch.NewMonitorSubscriber(getClientAddr(r), contentTypeParam)
		} else {
			s = dispatch.NewStandardSubscriber(getClientAddr(r), contentTypeParam)
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

func resolveContentType(r *http.Request) (string, error) {
	contentType := r.URL.Query().Get("type")
	if contentType == "" {
		return defaultContentType, nil
	}
	for _, t := range supportedContentTypes {
		if strings.ToLower(contentType) == strings.ToLower(t) {
			return contentType, nil
		}
	}
	return "", fmt.Errorf("The specified type (%s) is unsupported", contentType)
}
