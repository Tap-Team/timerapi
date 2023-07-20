package amidstr

import (
	"fmt"
	"strings"
)

func UnmarshalNullString(b []byte) (string, error) {
	v := string(b)
	v = strings.Trim(v, `"`)
	switch v {
	case "null":
		return "", nil
	default:
		return v, nil
	}
}

func MarshalNullString(s string) ([]byte, error) {
	if len(s) == 0 {
		return []byte("null"), nil
	}
	return []byte(`"` + s + `"`), nil
}

func ScanNullString(s *string, src interface{}) error {
	switch src := src.(type) {
	case nil:
		*s = ""
		return nil
	case string:
		*s = src
		return nil
	default:
		*s = ""
		return fmt.Errorf("cannot scan %T into *Email", src)
	}
}
