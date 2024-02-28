// http-utils.go - Provides utility methods for HTTP handling
package utils

import "net/http"

// Create a new HTTP request that contains the provided headers, method and url.
// Request has no body
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
