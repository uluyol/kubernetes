package main

import (
	"errors"
	"fmt"
	"time"
)

type parseTimeState int

const (
	sHour parseTimeState = iota + 1
	sMinute
	sSecond
	sUTC
	sOffHour
	sOffMinute
)

var parseTimeStateString = map[parseTimeState]string{
	sHour:      "hour",
	sMinute:    "minute",
	sSecond:    "second",
	sUTC:       "UTC",
	sOffHour:   "offset hour",
	sOffMinute: "offset minute",
}

type timeParseErr struct {
	state parseTimeState
}

func (t timeParseErr) Error() string {
	return "expected two digits for " + parseTimeStateString[t.state]
}

func getTwoDigits(s string) (int, bool) {
	if len(s) >= 2 && '0' <= s[0] && s[0] <= '9' && '0' <= s[1] && s[1] <= '9' {
		return int(s[0]-'0')*10 + int(s[1]-'0'), true
	}
	return 0, false
}

func zoneChar(b byte) bool {
	return b == 'Z' || b == '+' || b == '-'
}

func validate(x, min, max int, name string) error {
	if x < min || max < x {
		return fmt.Errorf("the %s must be within the range %d...%d", name, min, max)
	}
	return nil
}

type triState int

const (
	unset triState = iota
	setFalse
	setTrue
)

func parseTimeISO8601(s string) (time.Time, error) {
	theTime := struct {
		hour      int
		minute    int
		second    int
		utc       triState
		offNeg    bool
		offHour   int
		offMinute int
	}{}
	state := sHour
	isExtended := false
	for s != "" {
		switch state {
		case sHour:
			v, ok := getTwoDigits(s)
			if !ok {
				return time.Time{}, timeParseErr{state}
			}
			theTime.hour = v
			s = s[2:]
		case sMinute:
			if !zoneChar(s[0]) {
				if s[0] == ':' {
					isExtended = true
					s = s[1:]
				}
				v, ok := getTwoDigits(s)
				if !ok {
					return time.Time{}, timeParseErr{state}
				}
				theTime.minute = v
				s = s[2:]
			}
		case sSecond:
			if !zoneChar(s[0]) {
				if s[0] == ':' {
					if isExtended {
						s = s[1:]
					} else {
						return time.Time{}, errors.New("unexpected ':' before 'second' value")
					}
				} else if isExtended {
					return time.Time{}, errors.New("expected ':' before 'second' value")
				}
				v, ok := getTwoDigits(s)
				if !ok {
					return time.Time{}, timeParseErr{state}
				}
				theTime.second = v
				s = s[2:]
			}
		case sUTC:
			if s[0] == 'Z' {
				theTime.utc = setTrue
				s = s[1:]
			} else {
				theTime.utc = setFalse
			}
		case sOffHour:
			if theTime.utc == setTrue {
				return time.Time{}, errors.New("unexpected offset, already specified UTC")
			}
			var sign int
			if s[0] == '+' {
				sign = 1
			} else if s[0] == '-' {
				sign = -1
				theTime.offNeg = true
			} else {
				return time.Time{}, errors.New("offset must begin with '+' or '-'")
			}
			s = s[1:]
			v, ok := getTwoDigits(s)
			if !ok {
				return time.Time{}, timeParseErr{state}
			}
			theTime.offHour = sign * v
			s = s[2:]
		case sOffMinute:
			if s[0] == ':' {
				if isExtended {
					s = s[1:]
				} else {
					return time.Time{}, errors.New("unexpected ':' before 'minute' value")
				}
			} else if isExtended {
				return time.Time{}, errors.New("expected ':' before 'second' value")
			}
			v, ok := getTwoDigits(s)
			if !ok {
				return time.Time{}, timeParseErr{state}
			}
			theTime.offMinute = v
			s = s[2:]
		default:
			return time.Time{}, errors.New("an unknown error occured")
		}
		state++
	}
	if err := validate(theTime.hour, 0, 23, "hour"); err != nil {
		return time.Time{}, err
	}
	if err := validate(theTime.minute, 0, 59, "minute"); err != nil {
		return time.Time{}, err
	}
	if err := validate(theTime.second, 0, 59, "second"); err != nil {
		return time.Time{}, err
	}
	if err := validate(theTime.offHour, -12, 14, "offset hour"); err != nil {
		return time.Time{}, err
	}
	if err := validate(theTime.offMinute, 0, 59, "offset minute"); err != nil {
		return time.Time{}, err
	}
	if theTime.offNeg && theTime.offHour == 0 && theTime.offMinute == 0 {
		return time.Time{}, errors.New("an offset of -00 may not be used, must use +00")
	}
	var (
		loc *time.Location
		err error
	)
	if theTime.utc == setTrue {
		loc, err = time.LoadLocation("UTC")
		if err != nil {
			panic(err)
		}
	} else if theTime.utc == setFalse {
		loc = time.FixedZone("Zone", theTime.offMinute*60+theTime.offHour*3600)
	} else {
		loc, err = time.LoadLocation("Local")
		if err != nil {
			panic(err)
		}
	}
	t := time.Date(1, time.January, 1, theTime.hour, theTime.minute, theTime.second, 0, loc)
	return t, nil
}
