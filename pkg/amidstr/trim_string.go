package amidstr

import "strings"

func UnmarshalTrimString(s *string, b []byte) error {
	v := string(b)
	v = strings.Trim(v, `"`)
	v = strings.Trim(v, ` `)
	switch v {
	case "null":
		*s = ""
		return nil
	default:
		*s = v
		return nil
	}
}
