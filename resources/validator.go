package resources

import (
	log "github.com/Financial-Times/go-logger"
	"io"
	"io/ioutil"
	"net/http"
)

func isValidApiKey(providedApiKey string, masheryApiKeyValidationURL string, httpClient *http.Client) (bool, string, int) {
	if providedApiKey == "" {
		return false, "Empty api key", http.StatusUnauthorized
	}

	req, err := http.NewRequest("GET", masheryApiKeyValidationURL, nil)
	if err != nil {
		log.WithField("url", masheryApiKeyValidationURL).WithError(err).Error("Invalid URL for api key validation")
		return false, "Invalid URL", http.StatusInternalServerError
	}

	req.Header.Set(apiKeyHeaderField, providedApiKey)

	//if the api key has more than four characters we want to log the first four
	apiKeyFirstChars := ""
	if len(providedApiKey) > 4 {
		apiKeyFirstChars = providedApiKey[0:4]
	}
	log.WithField("url", req.URL.String()).WithField("apiKeyFirstChars", apiKeyFirstChars).Info("Calling Mashery to validate api key")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.WithField("url", req.URL.String()).WithError(err).Error("Cannot send request to Mashery")
		return false, "Request to validate api key failed", http.StatusInternalServerError
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	respStatusCode := resp.StatusCode
	if respStatusCode == http.StatusOK {
		return true, "", 0
	}

	if respStatusCode == http.StatusUnauthorized {
		log.WithField("apiKeyFirstChars", apiKeyFirstChars).Error("Invalid api key")
		return false, "Invalid api key", http.StatusUnauthorized
	}

	log.WithField("url", req.URL.String()).Errorf("Received unexpected status code from Mashery: %d", respStatusCode)
	return false, "Request to validate api key returned an unexpected response", http.StatusServiceUnavailable
}
