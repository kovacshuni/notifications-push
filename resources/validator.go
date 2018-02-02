package resources

import (
	log "github.com/Financial-Times/go-logger"
	"io"
	"io/ioutil"
	"net/http"
)

func isValidApiKey(providedApiKey string, apiGatewayApiKeyValidationURL string, httpClient *http.Client) (bool, string, int) {
	if providedApiKey == "" {
		return false, "Empty api key", http.StatusUnauthorized
	}

	req, err := http.NewRequest("GET", apiGatewayApiKeyValidationURL, nil)
	if err != nil {
		log.WithField("url", apiGatewayApiKeyValidationURL).WithError(err).Error("Invalid URL for api key validation")
		return false, "Invalid URL", http.StatusInternalServerError
	}

	req.Header.Set(apiKeyHeaderField, providedApiKey)

	//if the api key has more than 8 characters we want to log the first and last four
	apiKeyFirstChars := ""
	apiKeyLastChars := ""
	if len(providedApiKey) > 8 {
		apiKeyFirstChars = providedApiKey[:4]
		apiKeyLastChars = providedApiKey[len(providedApiKey)-4:]
	}
	log.WithField("url", req.URL.String()).WithField("apiKeyFirstChars", apiKeyFirstChars).WithField("apiKeyLastChars", apiKeyLastChars).Info("Calling Api Gateway to validate api key")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.WithField("url", req.URL.String()).WithError(err).Error("Cannot send request to Api Gateway")
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
		log.WithField("apiKeyFirstChars", apiKeyFirstChars).WithField("apiKeyLastChars", apiKeyLastChars).Error("Invalid api key")
		return false, "Invalid api key", http.StatusUnauthorized
	}

	log.WithField("url", req.URL.String()).WithField("apiKeyFirstChars", apiKeyFirstChars).WithField("apiKeyLastChars", apiKeyLastChars).Errorf("Received unexpected status code from Api Gateway: %d", respStatusCode)
	return false, "Request to validate api key returned an unexpected response", http.StatusServiceUnavailable
}
