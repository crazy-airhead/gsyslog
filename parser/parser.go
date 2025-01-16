package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	PriPartStart = '<'
	PriPartEnd   = '>'

	NoVersion = -1
)

var (
	ErrEOL     = &Error{Msg: "End of log line"}
	ErrNoSpace = &Error{Msg: "No space found"}

	ErrPriorityNoStart  = &Error{Msg: "No start char found for priority"}
	ErrPriorityEmpty    = &Error{Msg: "Priority field empty"}
	ErrPriorityNoEnd    = &Error{Msg: "No end char found for priority"}
	ErrPriorityTooShort = &Error{Msg: "Priority field too short"}
	ErrPriorityTooLong  = &Error{Msg: "Priority field too long"}
	ErrPriorityNonDigit = &Error{Msg: "Non digit found in priority"}

	ErrVersionNotFound = &Error{Msg: "Can not find version"}

	ErrTimestampUnknownFormat = &Error{Msg: "Timestamp codec unknown"}

	ErrHostnameTooShort = &Error{Msg: "Hostname field too short"}
)

type Parser interface {
	Parse(data []byte, client string) (*Log, error)
	Location(*time.Location)
}

type Error struct {
	Msg string
}

type Priority struct {
	P int
	F Facility
	S Severity
}

type Facility struct {
	Value int
}

type Severity struct {
	Value int
}

// ParsePriority https://tools.ietf.org/html/rfc3164#section-4.1
func ParsePriority(buff []byte, cursor *int, l int) (Priority, error) {
	pri := newPriority(0)

	if l <= 0 {
		return pri, ErrPriorityEmpty
	}

	if buff[*cursor] != PriPartStart {
		return pri, ErrPriorityNoStart
	}

	i := 1
	priDigit := 0

	for i < l {
		if i >= 5 {
			return pri, ErrPriorityTooLong
		}

		c := buff[i]

		if c == PriPartEnd {
			if i == 1 {
				return pri, ErrPriorityTooShort
			}

			*cursor = i + 1
			return newPriority(priDigit), nil
		}

		if IsDigit(c) {
			v, e := strconv.Atoi(string(c))
			if e != nil {
				return pri, e
			}

			priDigit = (priDigit * 10) + v
		} else {
			return pri, ErrPriorityNonDigit
		}

		i++
	}

	return pri, ErrPriorityNoEnd
}

// ParseVersion https://tools.ietf.org/html/rfc5424#section-6.2.2
func ParseVersion(buff []byte, cursor *int, l int) (int, error) {
	if *cursor >= l {
		return NoVersion, ErrVersionNotFound
	}

	c := buff[*cursor]
	*cursor++

	// XXX : not a version, not an error though as RFC 3164 does not support it
	if !IsDigit(c) {
		return NoVersion, nil
	}

	v, e := strconv.Atoi(string(c))
	if e != nil {
		*cursor--
		return NoVersion, e
	}

	return v, nil
}

func IsDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func newPriority(p int) Priority {
	// The Priority value is calculated by first multiplying the Facility
	// number by 8 and then adding the numerical value of the Severity.

	return Priority{
		P: p,
		F: Facility{Value: p / 8},
		S: Severity{Value: p % 8},
	}
}

func FindNextSpace(buff []byte, from int, l int) (int, error) {
	var to int

	for to = from; to < l; to++ {
		if buff[to] == ' ' {
			to++
			return to, nil
		}
	}

	return 0, ErrNoSpace
}

func Parse2Digits(buff []byte, cursor *int, l int, min int, max int, e error) (int, error) {
	digitLen := 2

	if *cursor+digitLen > l {
		return 0, ErrEOL
	}

	sub := string(buff[*cursor : *cursor+digitLen])

	*cursor += digitLen

	i, err := strconv.Atoi(sub)
	if err != nil {
		return 0, e
	}

	if i >= min && i <= max {
		return i, nil
	}

	return 0, e
}

func ParseHostname(buff []byte, cursor *int, l int) (string, error) {
	from := *cursor

	if from >= l {
		return "", ErrHostnameTooShort
	}

	var to int

	for to = from; to < l; to++ {
		if buff[to] == ' ' {
			break
		}
	}

	hostname := buff[from:to]

	*cursor = to

	return string(hostname), nil
}

func ShowCursorPos(buff []byte, cursor int) {
	fmt.Println(string(buff))
	padding := strings.Repeat("-", cursor)
	fmt.Println(padding + "↑\n")
}

func (err *Error) Error() string {
	return err.Msg
}
