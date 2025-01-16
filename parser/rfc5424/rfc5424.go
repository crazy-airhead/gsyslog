package rfc5424

import (
	"fmt"
	"github.com/crazy-airhead/gsyslog/parser"
	"math"
	"strconv"
	"time"
)

const (
	NilValue = '-'
)

var (
	ErrYearInvalid       = &parser.Error{Msg: "Invalid year in timestamp"}
	ErrMonthInvalid      = &parser.Error{Msg: "Invalid month in timestamp"}
	ErrDayInvalid        = &parser.Error{Msg: "Invalid day in timestamp"}
	ErrHourInvalid       = &parser.Error{Msg: "Invalid hour in timestamp"}
	ErrMinuteInvalid     = &parser.Error{Msg: "Invalid minute in timestamp"}
	ErrSecondInvalid     = &parser.Error{Msg: "Invalid second in timestamp"}
	ErrSecFracInvalid    = &parser.Error{Msg: "Invalid fraction of second in timestamp"}
	ErrTimeZoneInvalid   = &parser.Error{Msg: "Invalid time zone in timestamp"}
	ErrInvalidTimeFormat = &parser.Error{Msg: "Invalid time codec"}
	ErrInvalidAppName    = &parser.Error{Msg: "Invalid app name"}
	ErrInvalidProcId     = &parser.Error{Msg: "Invalid proc ID"}
	ErrInvalidMsgId      = &parser.Error{Msg: "Invalid msg ID"}
	ErrNoStructuredData  = &parser.Error{Msg: "No structured data"}
)

type Parser struct {
}

type partialTime struct {
	hour    int
	minute  int
	seconds int
	secFrac float64
}

type fullTime struct {
	pt  partialTime
	loc *time.Location
}

type fullDate struct {
	year  int
	month int
	day   int
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Location(location *time.Location) {
	// Ignore as RFC5424 syslog always has a timezone
}

func (p *Parser) Parse(data []byte, client string) (*parser.Log, error) {
	log := parser.NewLog(data)
	err := p.parseHeader(log)
	if err != nil {
		log.Err = err
		return log, err
	}

	err = p.parseStructuredData(log)
	if err != nil {
		log.Err = err
		return log, err
	}

	log.MoveCursor()

	if log.Cursor() < log.Len() {
		log.SetMessage(string(log.Body[log.Cursor():]))
	} else {
		log.SetMessage("")
	}

	return log, nil
}

// HEADER = PRI VERSION SP TIMESTAMP SP HOSTNAME SP APP-NAME SP PROCID SP MSGID
func (p *Parser) parseHeader(log *parser.Log) error {
	err := parsePriority(log)
	if err != nil {
		return err
	}

	err = parseVersion(log)
	if err != nil {
		return err
	}

	// move over a blank
	log.MoveCursor()

	err = p.parseTimestamp(log)
	if err != nil {
		return err
	}

	// move over a blank
	log.MoveCursor()

	err = parseHostname(log)
	if err != nil {
		return err
	}

	// move over a blank
	log.MoveCursor()

	err = parseAppName(log)
	if err != nil {
		return err
	}

	// move over a blank
	log.MoveCursor()

	err = parseProcId(log)
	if err != nil {
		return nil
	}

	// move over a blank
	log.MoveCursor()

	err = parseMsgId(log)
	if err != nil {
		return nil
	}

	// move over a blank
	log.MoveCursor()

	return nil
}

func parsePriority(log *parser.Log) error {
	cursor := log.Cursor()
	priority, err := parser.ParsePriority(log.Body, &cursor, log.Len())
	if err != nil {
		// RFC3164 sec 4.3.3
		log.SetPriority(13)
		log.SetFacility(1)
		log.SetSeverity(5)

		return err
	}

	log.SetPriority(priority.P)
	log.SetFacility(priority.F.Value)
	log.SetSeverity(priority.S.Value)

	// 成功移动光标
	log.SetCursor(cursor)

	return nil
}

func parseVersion(log *parser.Log) error {
	cursor := log.Cursor()
	version, err := parser.ParseVersion(log.Body, &cursor, log.Len())
	if err != nil {
		return err
	}

	log.SetVersion(version)

	// 成功移动光标
	log.SetCursor(cursor)

	return nil
}

// https://tools.ietf.org/html/rfc5424#section-6.2.3
func (p *Parser) parseTimestamp(log *parser.Log) error {
	var ts time.Time

	cursor := log.Cursor()
	if cursor >= log.Len() {
		return ErrInvalidTimeFormat
	}

	if log.Body[cursor] == NilValue {
		log.MoveCursor()
		return nil
	}

	fd, err := parseFullDate(log.Body, &cursor, log.Len())
	if err != nil {
		return err
	}

	if cursor >= log.Len() || log.Body[cursor] != 'T' {
		log.SetCursor(cursor)
		return ErrInvalidTimeFormat
	}

	cursor++

	ft, err := parseFullTime(log.Body, &cursor, log.Len())
	if err != nil {
		log.SetCursor(cursor)
		return parser.ErrTimestampUnknownFormat
	}

	nSec, err := toNSec(ft.pt.secFrac)
	if err != nil {
		log.SetCursor(cursor)
		return err
	}

	ts = time.Date(
		fd.year,
		time.Month(fd.month),
		fd.day,
		ft.pt.hour,
		ft.pt.minute,
		ft.pt.seconds,
		nSec,
		ft.loc,
	)

	log.SetTimestamp(ts)

	//成功，设置光标
	log.SetCursor(cursor)

	return nil
}

// HOSTNAME = NilValue / 1*255PRINTUSASCII
func parseHostname(log *parser.Log) error {
	cursor := log.Cursor()
	hostname, err := parser.ParseHostname(log.Body, &cursor, log.Len())
	if err != nil {
		return err
	}

	log.SetHostname(hostname)
	log.SetCursor(cursor)
	return nil
}

// APP-NAME = NilValue / 1*48PRINTUSASCII
func parseAppName(log *parser.Log) error {
	cursor := log.Cursor()
	appName, err := parseUpToLen(log.Body, &cursor, log.Len(), 48, ErrInvalidAppName)
	if err != nil {
		return err
	}

	log.SetAppName(appName)
	log.SetCursor(cursor)
	return nil
}

// PROCID = NilValue / 1*128PRINTUSASCII
func parseProcId(log *parser.Log) error {
	cursor := log.Cursor()
	procId, err := parseUpToLen(log.Body, &cursor, log.Len(), 128, ErrInvalidAppName)
	if err != nil {
		return err
	}

	log.SetProcId(procId)
	log.SetCursor(cursor)
	return nil
}

// MSGID = NilValue / 1*32PRINTUSASCII
func parseMsgId(log *parser.Log) error {
	cursor := log.Cursor()
	msgId, err := parseUpToLen(log.Body, &cursor, log.Len(), 32, ErrInvalidAppName)
	if err != nil {
		return err
	}

	log.SetMsgId(msgId)
	log.SetCursor(cursor)
	return nil
}

func (p *Parser) parseStructuredData(log *parser.Log) error {
	cursor := log.Cursor()
	sd, err := parseStructuredData(log.Body, &cursor, log.Len())
	if err != nil {
		return err
	}

	log.SetStructuredData(sd)
	log.SetCursor(cursor)
	return nil
}

// FULL-DATE : DATE-FULLYEAR "-" DATE-MONTH "-" DATE-MDAY
func parseFullDate(buff []byte, cursor *int, l int) (fullDate, error) {
	var fd fullDate

	year, err := parseYear(buff, cursor, l)
	if err != nil {
		return fd, err
	}

	if *cursor >= l || buff[*cursor] != '-' {
		return fd, parser.ErrTimestampUnknownFormat
	}

	*cursor++

	month, err := parseMonth(buff, cursor, l)
	if err != nil {
		return fd, err
	}

	if *cursor >= l || buff[*cursor] != '-' {
		return fd, parser.ErrTimestampUnknownFormat
	}

	*cursor++

	day, err := parseDay(buff, cursor, l)
	if err != nil {
		return fd, err
	}

	fd = fullDate{
		year:  year,
		month: month,
		day:   day,
	}

	return fd, nil
}

// DATE-FULLYEAR   = 4DIGIT
func parseYear(buff []byte, cursor *int, l int) (int, error) {
	yearLen := 4

	if *cursor+yearLen > l {
		return 0, parser.ErrEOL
	}

	// XXX : we do not check for a valid year (ie. 1999, 2013 etc)
	// XXX : we only checks the format is correct
	sub := string(buff[*cursor : *cursor+yearLen])

	*cursor += yearLen

	year, err := strconv.Atoi(sub)
	if err != nil {
		return 0, ErrYearInvalid
	}

	return year, nil
}

// DATE-MONTH = 2DIGIT  ; 01-12
func parseMonth(buff []byte, cursor *int, l int) (int, error) {
	return parser.Parse2Digits(buff, cursor, l, 1, 12, ErrMonthInvalid)
}

// DATE-MDAY = 2DIGIT  ; 01-28, 01-29, 01-30, 01-31 based on month/year
func parseDay(buff []byte, cursor *int, l int) (int, error) {
	// XXX : this is a relaxed constraint
	// XXX : we do not check if valid regarding February or leap years
	// XXX : we only checks that day is in range [01 -> 31]
	// XXX : in other words this function will not rant if you provide Feb 31th
	return parser.Parse2Digits(buff, cursor, l, 1, 31, ErrDayInvalid)
}

// FULL-TIME = PARTIAL-TIME TIME-OFFSET
func parseFullTime(buff []byte, cursor *int, l int) (fullTime, error) {
	var loc = new(time.Location)
	var ft fullTime

	pt, err := parsePartialTime(buff, cursor, l)
	if err != nil {
		return ft, err
	}

	loc, err = parseTimeOffset(buff, cursor, l)
	if err != nil {
		return ft, err
	}

	ft = fullTime{
		pt:  pt,
		loc: loc,
	}

	return ft, nil
}

// PARTIAL-TIME = TIME-HOUR ":" TIME-MINUTE ":" TIME-SECOND[TIME-SECFRAC]
func parsePartialTime(buff []byte, cursor *int, l int) (partialTime, error) {
	var pt partialTime

	hour, minute, err := getHourMinute(buff, cursor, l)
	if err != nil {
		return pt, err
	}

	if *cursor >= l || buff[*cursor] != ':' {
		return pt, ErrInvalidTimeFormat
	}

	*cursor++

	// ----

	seconds, err := parseSecond(buff, cursor, l)
	if err != nil {
		return pt, err
	}

	pt = partialTime{
		hour:    hour,
		minute:  minute,
		seconds: seconds,
	}

	// ----

	if *cursor >= l || buff[*cursor] != '.' {
		return pt, nil
	}

	*cursor++

	secFrac, err := parseSecFrac(buff, cursor, l)
	if err != nil {
		return pt, nil
	}
	pt.secFrac = secFrac

	return pt, nil
}

// TIME-HOUR = 2DIGIT  ; 00-23
func parseHour(buff []byte, cursor *int, l int) (int, error) {
	return parser.Parse2Digits(buff, cursor, l, 0, 23, ErrHourInvalid)
}

// TIME-MINUTE = 2DIGIT  ; 00-59
func parseMinute(buff []byte, cursor *int, l int) (int, error) {
	return parser.Parse2Digits(buff, cursor, l, 0, 59, ErrMinuteInvalid)
}

// TIME-SECOND = 2DIGIT  ; 00-59
func parseSecond(buff []byte, cursor *int, l int) (int, error) {
	return parser.Parse2Digits(buff, cursor, l, 0, 59, ErrSecondInvalid)
}

// TIME-SECFRAC = "." 1*6DIGIT
func parseSecFrac(buff []byte, cursor *int, l int) (float64, error) {
	maxDigitLen := 6

	max := *cursor + maxDigitLen
	from := *cursor
	to := from

	for to = from; to < max; to++ {
		if to >= l {
			break
		}

		c := buff[to]
		if !parser.IsDigit(c) {
			break
		}
	}

	sub := string(buff[from:to])
	if len(sub) == 0 {
		return 0, ErrSecFracInvalid
	}

	secFrac, err := strconv.ParseFloat("0."+sub, 64)
	*cursor = to
	if err != nil {
		return 0, ErrSecFracInvalid
	}

	return secFrac, nil
}

// TIME-OFFSET = "Z" / TIME-NUMOFFSET
func parseTimeOffset(buff []byte, cursor *int, l int) (*time.Location, error) {

	if *cursor >= l || buff[*cursor] == 'Z' {
		*cursor++
		return time.UTC, nil
	}

	return parseNumericalTimeOffset(buff, cursor, l)
}

// TIME-NUMOFFSET  = ("+" / "-") TIME-HOUR ":" TIME-MINUTE
func parseNumericalTimeOffset(buff []byte, cursor *int, l int) (*time.Location, error) {
	var loc = new(time.Location)

	sign := buff[*cursor]

	if (sign != '+') && (sign != '-') {
		return loc, ErrTimeZoneInvalid
	}

	*cursor++

	hour, minute, err := getHourMinute(buff, cursor, l)
	if err != nil {
		return loc, err
	}

	tzStr := fmt.Sprintf("%s%02d:%02d", string(sign), hour, minute)
	tmpTs, err := time.Parse("-07:00", tzStr)
	if err != nil {
		return loc, err
	}

	return tmpTs.Location(), nil
}

func getHourMinute(buff []byte, cursor *int, l int) (int, int, error) {
	hour, err := parseHour(buff, cursor, l)
	if err != nil {
		return 0, 0, err
	}

	if *cursor >= l || buff[*cursor] != ':' {
		return 0, 0, ErrInvalidTimeFormat
	}

	*cursor++

	minute, err := parseMinute(buff, cursor, l)
	if err != nil {
		return 0, 0, err
	}

	return hour, minute, nil
}

func toNSec(sec float64) (int, error) {
	_, frac := math.Modf(sec)
	fracStr := strconv.FormatFloat(frac, 'f', 9, 64)
	fracInt, err := strconv.Atoi(fracStr[2:])
	if err != nil {
		return 0, err
	}

	return fracInt, nil
}

// ------------------------------------------------
// https://tools.ietf.org/html/rfc5424#section-6.3
// ------------------------------------------------

func parseStructuredData(buff []byte, cursor *int, l int) (string, error) {
	var sdData string
	var found bool

	if *cursor >= l {
		return "-", nil
	}

	if buff[*cursor] == NilValue {
		*cursor++
		return "-", nil
	}

	if buff[*cursor] != '[' {
		return sdData, ErrNoStructuredData
	}

	from := *cursor
	to := from

	for to = from; to < l; to++ {
		if found {
			break
		}

		b := buff[to]

		if b == ']' {
			switch t := to + 1; {
			case t == l:
				found = true
			case t <= l && buff[t] == ' ':
				found = true
			}
		}
	}

	if found {
		*cursor = to
		return string(buff[from:to]), nil
	}

	return sdData, ErrNoStructuredData
}

func parseUpToLen(buff []byte, cursor *int, l int, maxLen int, e error) (string, error) {
	var to int
	var found bool
	var result string

	max := *cursor + maxLen

	for to = *cursor; (to <= max) && (to < l); to++ {
		if buff[to] == ' ' {
			found = true
			break
		}
	}

	if found {
		result = string(buff[*cursor:to])
	} else if to > max {
		to = max // don't go past max
	}

	*cursor = to

	if found {
		return result, nil
	}

	return "", e
}
