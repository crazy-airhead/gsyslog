package codec

import (
	"bytes"
	"github.com/crazy-airhead/gsyslog/parser"
	"github.com/crazy-airhead/gsyslog/parser/rfc5424"
	"github.com/panjf2000/gnet/v2"
	"strconv"
)

type RFC6587Codec struct{}

func (f *RFC6587Codec) GetParser(data []byte) parser.Parser {
	return rfc5424.NewParser()
}

func (f *RFC6587Codec) Decode(conn gnet.Conn) ([]byte, error) {
	return nil, nil
}

func rfc6587ScannerSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, ' '); i > 0 {
		pLength := data[0:i]
		length, err := strconv.Atoi(string(pLength))
		if err != nil {
			if string(data[0:1]) == "<" {
				// Assume this frame uses non-transparent-framing
				return len(data), data, nil
			}
			return 0, nil, err
		}
		end := length + i + 1
		if len(data) >= end {
			// Return the frame with the length removed
			return end, data[i+1 : end], nil
		}
	}

	// Request more data
	return 0, nil, nil
}
