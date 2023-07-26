package amidtime

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

// swagger:type int
type DateTime time.Time

func (t DateTime) T() time.Time {
	return time.Time(t)
}

func (t DateTime) Date() Date {
	return Date(t)
}

func (t DateTime) String() string {
	return t.T().Format(time.DateTime)
}

func (t *DateTime) Value() (driver.Value, error) {
	tm := t.T()
	if tm.IsZero() {
		return nil, nil
	}
	return tm, nil
}

func (t *DateTime) Scan(src any) error {
	switch src := src.(type) {
	case nil:
		*t = DateTime{}
		return nil
	case int64:
		tm := time.Unix(src, 0)
		*t = DateTime(tm)
	case time.Time:
		*t = DateTime(src)
		return nil
	}
	return fmt.Errorf("cannot scan %T", src)
}

func (t *DateTime) UnmarshalJSON(b []byte) error {
	value, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}
	tm := time.Unix(value, 0)
	*t = DateTime(tm)
	return nil
}

func (t DateTime) MarshalJSON() ([]byte, error) {
	u := t.T().Unix()
	if t.T().IsZero() {
		u = 0
	}
	return []byte(fmt.Sprint(u)), nil
}

func (t DateTime) Unix() int64 {
	return t.T().Unix()
}

func Now() DateTime {
	return DateTime(time.Now())
}
