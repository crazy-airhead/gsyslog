package rfc3164

import (
	"bytes"
	"errors"
	"github.com/crazy-airhead/gsyslog/parser"
	"os"
	"strings"
	"time"
)

type Parser struct {
	location *time.Location
}

func NewParser() *Parser {
	return &Parser{
		location: time.UTC,
	}
}

func (p *Parser) Location(location *time.Location) {
	p.location = location
}

func (p *Parser) Parse(data []byte, client string) (*parser.Log, error) {
	log := parser.NewLog(data)
	log.SetClient(client)

	err := parsePriority(log)
	if err != nil {
		log.SetTimestamp(time.Now().Round(time.Second))
		log.SetHostname("")

		log.SetTag("")
		err = parseContent(log)
		log.Err = err
		return log, err
	}

	tCursor := log.Cursor()
	err = p.parseHeader(log)
	if errors.Is(err, parser.ErrTimestampUnknownFormat) {
		// RFC3164 sec 4.3.2.
		log.SetTimestamp(time.Now().Round(time.Second))
		log.SetHostname("")

		// No tag processing should be done
		log.SetSkipTag(true)
		// Reset cursor for content read
		log.SetCursor(tCursor)
	} else if err != nil {
		log.Err = err
		return log, nil
	} else {
		log.MoveCursor()
	}

	err = parseMessage(log)
	if !errors.Is(err, parser.ErrEOL) {
		log.Err = err
		return log, err
	}

	return log, nil
}

// parsePriority 识别成功时移动光标，未成功时不移动
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

func (p *Parser) parseHeader(log *parser.Log) error {
	err := parseTimestamp(log, p.location)
	if err != nil {
		return err
	}

	err = parseHostname(log)
	if err != nil {
		return err
	}

	return nil
}

// https://tools.ietf.org/html/rfc3164#section-4.1.2
func parseTimestamp(log *parser.Log, location *time.Location) error {
	var ts time.Time
	var err error
	var tsFmtLen int
	var sub []byte

	tsFmts := []string{
		time.Stamp,
		time.RFC3339,
	}
	// if timestamps starts with numeric try formats with different order
	// it is more likely that timestamp is in RFC3339 format then
	cursor := log.Cursor()
	if c := log.Body[cursor]; c > '0' && c < '9' {
		tsFmts = []string{
			time.RFC3339,
			time.Stamp,
		}
	}

	found := false
	for _, tsFmt := range tsFmts {
		tsFmtLen = len(tsFmt)

		if cursor+tsFmtLen > log.Len() {
			continue
		}

		sub = log.Body[cursor : tsFmtLen+cursor]
		ts, err = time.ParseInLocation(tsFmt, string(sub), location)
		if err == nil {
			found = true
			break
		}
	}

	if !found {
		cursor = len(time.Stamp)

		// XXX : If the timestamp is invalid we try to push the cursor one byte
		// XXX : further, in case it is a space
		if (cursor < log.Len()) && (log.Body[cursor] == ' ') {
			cursor++
		}

		log.SetCursor(cursor)
		log.SetTimestamp(ts)
		return parser.ErrTimestampUnknownFormat
	}

	fixTimestampIfNeeded(&ts)

	cursor += tsFmtLen

	if (cursor < log.Len()) && (log.Body[cursor] == ' ') {
		cursor++
	}

	log.SetCursor(cursor)
	log.SetTimestamp(ts)
	return nil
}

func parseHostname(log *parser.Log) error {
	cursor := log.Cursor()
	hostname, err := parser.ParseHostname(log.Body, &cursor, log.Len())
	if err == nil && len(hostname) > 0 && string(hostname[len(hostname)-1]) == ":" { // not an hostname! we found a GNU implementation of syslog()
		log.MoveCursorN(-1)
		hostname, err = os.Hostname()
		if err == nil {
			log.SetHostname(hostname)
			return nil
		}

		log.SetHostname("")
		fixHostname(log)
		return nil
	}

	log.SetHostname(hostname)
	fixHostname(log)
	log.SetCursor(cursor)
	return err
}

func parseMessage(log *parser.Log) error {
	if !log.SkipTag() {
		err := parseTag(log)
		if err != nil {
			return err
		}
	} else {
		log.SetTag("")
	}

	err := parseContent(log)
	if !errors.Is(err, parser.ErrEOL) {
		return err
	}

	return err
}

// http://tools.ietf.org/html/rfc3164#section-4.1.3
func parseTag(log *parser.Log) error {
	var b byte
	var endOfTag bool
	var bracketOpen bool
	var tag []byte
	var err error
	var found bool

	from := log.Cursor()
	cursor := log.Cursor()
	for {
		if cursor == log.Len() {
			log.SetTag("")
			return nil
		}

		b = log.Body[cursor]
		bracketOpen = b == '['
		endOfTag = b == ':' || b == ' '

		// XXX : parse PID ?
		if bracketOpen {
			tag = log.Body[from:cursor]
			found = true
		}

		if endOfTag {
			if !found {
				tag = log.Body[from:cursor]
				found = true
			}

			cursor++
			break
		}

		cursor++
	}

	if (cursor < log.Len()) && (log.Body[cursor] == ' ') {
		cursor++
	}

	log.SetTag(string(tag))
	log.SetCursor(cursor)

	return err
}

func parseContent(log *parser.Log) error {
	if log.Cursor() > log.Len() {
		log.SetContent("")
		return parser.ErrEOL
	}

	content := bytes.Trim(log.Body[log.Cursor():log.Len()], " ")
	log.MoveCursorN(len(content))

	log.SetContent(string(content))
	return nil
}

func fixTimestampIfNeeded(ts *time.Time) {
	now := time.Now()
	y := ts.Year()

	if ts.Year() == 0 {
		y = now.Year()
	}

	newTs := time.Date(y, ts.Month(), ts.Day(), ts.Hour(), ts.Minute(),
		ts.Second(), ts.Nanosecond(), ts.Location())

	*ts = newTs
}

func fixHostname(log *parser.Log) {
	hostname := log.GetString("hostname")
	if hostname != "" {
		return
	}

	client := log.GetString("client")
	if i := strings.Index(client, ":"); i > 1 {
		log.SetHostname(client[:i])
	}

	log.SetHostname(client)
}
