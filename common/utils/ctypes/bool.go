package ctypes

import (
	"encoding/json"
	"strings"
)

type StrongBool bool

func (sb StrongBool) Value() bool {
	return bool(sb)
}

func (sb StrongBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(bool(sb))
}

func (sb *StrongBool) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*sb = parseToStrongBool(v)
	return nil
}

func parseToStrongBool(val interface{}) StrongBool {
	switch v := val.(type) {
	case bool:
		return StrongBool(v)
	case string:
		s := strings.ToLower(strings.TrimSpace(v))
		return StrongBool(!(s == "false" || s == "0" || s == ""))
	case float64:
		return StrongBool(v != 0)
	case nil:
		return false
	default:
		return false
	}
}
