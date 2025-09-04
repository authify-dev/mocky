package validator_controller

import "fmt"

type ValidationError struct {
	Path string
	Err  string
}

func (e ValidationError) String() string {
	if e.Path == "" {
		return e.Err
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Err)
}
