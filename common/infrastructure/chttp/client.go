package chttp

import "net/http"

type CustomApiClient struct {
	base_url       string
	api_key        string
	http_client    *http.Client
	api_key_header string
}

func NewCustomApiClient(base_url, api_key, api_key_header string) *CustomApiClient {

	if api_key_header == "" {
		api_key_header = "x-api-key"
	}

	return &CustomApiClient{
		base_url:       base_url,
		api_key:        api_key,
		http_client:    http.DefaultClient,
		api_key_header: api_key_header,
	}
}
