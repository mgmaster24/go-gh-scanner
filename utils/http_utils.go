package utils

import "net/http"

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
