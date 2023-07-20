package amidtime

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	ErrWrongDateFormat    = errors.New("wrong date format")
	ErrWrongCsvTimeFormat = errors.New("wrong date format, date should be in 'dd.MM.YYYY pattern'")
	ErrWrongCsvDay        = errors.New("wrong day format, day should be in 'dd' format")
	ErrWrongCsvMonth      = errors.New("wrong month format, day should be in 'MM' format")
	ErrWrongCsvYear       = errors.New("wrong year format, year should be in 'YYYY' format")
)

type Date time.Time

func NewDate(year int, month int, day int) Date {
	return Date(time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Now().Local().Location()))
}

func (d Date) Parse(s string, r *reflect.Value) {
	t := ParseCsvDate(s)
	r.Set(reflect.ValueOf(t))
}

func ParseCsvDate(s string) Date {
	if ok, _ := IsCsvDate(s); ok {
		dmy := strings.Split(s, ".")
		d, _ := strconv.Atoi(dmy[0])
		m, _ := strconv.Atoi(dmy[1])
		y, _ := strconv.Atoi(dmy[2])
		t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
		return Date(t)
	}
	return *new(Date)
}

func IsCsvDate(s string) (bool, error) {
	dmy := strings.Split(s, ".")
	if len(dmy) != 3 {
		return false, ErrWrongCsvTimeFormat
	}
	if len(dmy[0]) != 2 {
		return false, ErrWrongCsvDay
	}
	if len(dmy[1]) != 2 {
		return false, ErrWrongCsvMonth
	}
	if len(dmy[2]) != 4 {
		return false, ErrWrongCsvYear
	}
	return true, nil
}

func (t Date) T() time.Time {
	return time.Time(t)
}

func (t Date) String() string {
	return t.T().Format(time.DateOnly)
}

// returns string in dd.mm.yyyy format
func (t Date) Human() string {
	time := t.T()
	return fmt.Sprintf("%02d.%02d.%d", time.Day(), time.Month(), time.Year())
}

func (t *Date) Scan(src any) error {
	switch src := src.(type) {
	case nil:
		*t = Date{}
		return nil
	case string:
		tm, err := time.Parse(time.DateOnly, src)
		if err != nil {
			return ErrWrongDateFormat
		}
		*t = Date(tm)
		return nil
	case time.Time:
		*t = Date(src)
		return nil
	}

	return fmt.Errorf("cannot scan %T", src)
}

func (t *Date) UnmarshalJSON(b []byte) error {
	value := strings.Trim(string(b), `"`)
	if value == "" || value == "null" {
		return nil
	}

	tm, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return ErrWrongDateFormat
	}
	*t = Date(tm)
	return nil
}

func (t Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.T().Format(time.DateOnly) + `"`), nil
}
