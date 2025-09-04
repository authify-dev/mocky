package chttp

import "net/http"

func (c *CustomApiClient) setHeaders(req *http.Request, headers map[string]string) {
	req.Header.Set(c.api_key_header, c.api_key)

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}
