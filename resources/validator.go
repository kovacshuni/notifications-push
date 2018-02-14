package resources

import (
	log "github.com/Financial-Times/go-logger"
	"io"
	"io/ioutil"
	"net/http"
)

func isValidApiKey(providedApiKey string, apiGatewayKeyValidationURL string, httpClient *http.Client) (bool, string, int) {
	if providedApiKey == "" {
		return false, "Empty api key", http.StatusUnauthorized
	}

	req, err := http.NewRequest("GET", apiGatewayKeyValidationURL, nil)
	if err != nil {
		log.WithField("url", apiGatewayKeyValidationURL).WithError(err).Error("Invalid URL for api key validation")
		return false, "Invalid URL", http.StatusInternalServerError
	}

	req.Header.Set(apiKeyHeaderField, providedApiKey)

	//if the api key has more than five characters we want to log the last five
	keySuffix := ""
	if len(providedApiKey) > 5 {
		keySuffix = providedApiKey[len(providedApiKey)-5:]
	}
	log.WithField("url", req.URL.String()).WithField("keySuffix", keySuffix).Info("Calling the API Gateway to validate api key")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.WithField("url", req.URL.String()).WithError(err).Error("Cannot send request to the API Gateway")
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
		log.WithField("keySuffix", keySuffix).Error("Invalid api key")
		return false, "Invalid api key", http.StatusUnauthorized
	}

	if respStatusCode == http.StatusTooManyRequests {
		log.WithField("keySuffix", keySuffix).Error("API key rate limit exceeded.")
		return false, "Rate limit exceeded", http.StatusTooManyRequests
	}

	if respStatusCode == http.StatusForbidden {
		log.WithField("keySuffix", keySuffix).Error("Operation forbidden.")
		return false, "Operation forbidden", http.StatusForbidden
	}

	log.WithField("url", req.URL.String()).Errorf("Received unexpected status code from the API Gateway: %d", respStatusCode)
	return false, "Request to validate api key returned an unexpected response", http.StatusServiceUnavailable
}
