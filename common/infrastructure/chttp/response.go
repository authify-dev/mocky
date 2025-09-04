package chttp

type CustomApiResponse struct {
	Data       map[string]any `json:"data"`
	StatusCode string         `json:"status_code"`
}
