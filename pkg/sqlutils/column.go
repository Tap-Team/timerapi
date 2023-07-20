package sqlutils

import (
	"fmt"
	"strings"
)

type Column interface {
	Table() string
	String() string
}

// split columns with ',' sep
func Full(c ...Column) string {
	if len(c) == 0 {
		return ""
	}
	if len(c) == 1 {
		column := c[0]
		return fmt.Sprintf("%s.%s", column.Table(), column.String())
	}
	s := new(strings.Builder)
	for i, column := range c {
		s.WriteString(fmt.Sprintf("%s.%s", column.Table(), column.String()))
		if i < len(c)-1 {
			s.WriteRune(',')
		}
	}
	return s.String()
}
