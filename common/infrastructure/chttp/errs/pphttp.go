package chttp_errs

import "common/utils/cerrs"

type CustomHTTPError struct {
	*cerrs.CustomError
}
