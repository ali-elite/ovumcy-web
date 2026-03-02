package services

import (
	"errors"
	"time"
)

var (
	ErrDayRangeFromInvalid = errors.New("day range invalid from")
	ErrDayRangeToInvalid   = errors.New("day range invalid to")
	ErrDayRangeInvalid     = errors.New("day range invalid")
)

func ParseDayRange(rawFrom string, rawTo string, location *time.Location) (time.Time, time.Time, error) {
	from, err := ParseDayDate(rawFrom, location)
	if err != nil {
		return time.Time{}, time.Time{}, ErrDayRangeFromInvalid
	}
	to, err := ParseDayDate(rawTo, location)
	if err != nil {
		return time.Time{}, time.Time{}, ErrDayRangeToInvalid
	}
	if to.Before(from) {
		return time.Time{}, time.Time{}, ErrDayRangeInvalid
	}
	return from, to, nil
}
