package utils

import (
	"common/utils/cerrs"
)

// --------------------------------
// UTILS
// --------------------------------
// Utils
//--------------------------------

type Result[R any] struct {
	Data R
	Err  cerrs.CustomErrorInterface
}
