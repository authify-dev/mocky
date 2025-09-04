package chttp_errs

import (
	"common/utils/cerrs"
	"strconv"
)

type ReadingResponseError struct {
	CustomHTTPError
}

func NewReadingResponseError(options RequestOptions) *ReadingResponseError {
	return &ReadingResponseError{
		CustomHTTPError: CustomHTTPError{
			CustomError: &cerrs.CustomError{
				Code:    500,
				Message: "error reading the response from " + options.URL + "." + options.Method,
				Scope:   "konectus.knhttp.ReadingResponseError",
			},
		},
	}
}

type ExpectedCodeOptions struct {
	ExpectedCode int
	ActualCode   int
}

type NoExpectedCodeError struct {
	CustomHTTPError
}

func NewNoExpectedCodeError(options ExpectedCodeOptions) *NoExpectedCodeError {
	return &NoExpectedCodeError{
		CustomHTTPError: CustomHTTPError{
			CustomError: &cerrs.CustomError{
				Code:    500,
				Message: "expected code " + strconv.Itoa(options.ExpectedCode) + " but got " + strconv.Itoa(options.ActualCode),
				Scope:   "konectus.knhttp.NoExpectedCodeError",
			},
		},
	}
}
