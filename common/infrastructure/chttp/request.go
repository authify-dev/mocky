package chttp

type CustomApiRequest struct {
	Path         string
	Payload      any
	Headers      map[string]string
	ExpectedCode int  `default:"200"`
	Debug        bool `default:"false"`
}
