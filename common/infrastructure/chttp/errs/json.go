package chttp_errs

import "common/utils/cerrs"

type MarshalError struct {
	CustomHTTPError
}

func NewMarshalError() *MarshalError {
	return &MarshalError{
		CustomHTTPError: CustomHTTPError{
			CustomError: &cerrs.CustomError{
				Code:    500,
				Message: "error marshaling the payload to JSON",
				Scope:   "konectus.knhttp.MarshalError",
			},
		},
	}
}

type UnmarshalError struct {
	CustomHTTPError
}

func NewUnmarshalError() *UnmarshalError {
	return &UnmarshalError{
		CustomHTTPError: CustomHTTPError{
			CustomError: &cerrs.CustomError{
				Code:    500,
				Message: "error unmarshaling the response from JSON",
				Scope:   "konectus.knhttp.UnmarshalError",
			},
		},
	}
}
