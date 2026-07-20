// http-utils.go - Provides utility methods for HTTP handling
package utils

import (
	"net/http"
	"time"
)

// NewHttpRequestNoBody creates an HTTP request with no body and the provided headers.
func NewHttpRequestNoBody(httpMethod string, url string, headers *map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(httpMethod, url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range *headers {
		req.Header.Add(k, v)
	}

	return req, nil
}

// Retry calls fn up to maxAttempts times, doubling the delay after each
// failure. Only retries on non-nil errors; the caller is responsible for
// deciding whether the error is retryable (e.g. network errors, 5xx).
func Retry(maxAttempts int, initialDelay time.Duration, fn func() error) error {
	delay := initialDelay
	var err error
	for i := range maxAttempts {
		if err = fn(); err == nil {
			return nil
		}
		if i < maxAttempts-1 {
			time.Sleep(delay)
			delay *= 2
		}
	}
	return err
}
