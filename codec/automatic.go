package codec

import (
	"bytes"
	"errors"
	"github.com/crazy-airhead/gsyslog/parser"
	"github.com/crazy-airhead/gsyslog/parser/rfc3164"
	"github.com/crazy-airhead/gsyslog/parser/rfc5424"
	"github.com/panjf2000/gnet/v2"
	"strconv"
)

/* Selecting an 'AutomaticCodec' codec detects incoming codec (i.e. RFC3164 vs RFC5424) and Framing
 * (i.e. RFC6587 s3.4.1 octet counting as described here as RFC6587, and either no framing or
 * RFC6587 s3.4.2 octet stuffing / non-transparent framing, described here as either RFC3164
 * or RFC6587).
 *
 * In essence if you don't know which codec to select, or have multiple incoming formats, this
 * is the one to go for. There is a theoretical performance penalty (it has to look at a few bytes
 * at the start of the frame), and a risk that you may parse things you don't want to parse
 * (rogue syslog clients using other formats), so if you can be absolutely sure of your syslog
 * codec, it would be best to select it explicitly.
 */

type AutomaticCodec struct{}

const (
	Unknown = iota
	RFC3164 = iota
	RFC5424 = iota
	RFC6587 = iota
)

var (
	// 解析器
	rfc3164Parser = rfc3164.NewParser() // RFC3164: http://www.ietf.org/rfc/rfc3164.txt
	rfc5424Parser = rfc5424.NewParser() // RFC5424: http://www.ietf.org/rfc/rfc5424.txt
	rfc6587Parser = rfc5424.NewParser() // RFC5424: http://www.ietf.org/rfc/rfc6587.txt

	rfc3164Codec = &RFC3164Codec{} // RFC3164: http://www.ietf.org/rfc/rfc3164.txt
	rfc5424Codec = &RFC5424Codec{} // RFC5424: http://www.ietf.org/rfc/rfc5424.txt
	rfc6587Codec = &RFC6587Codec{} // RFC6587: http://www.ietf.org/rfc/rfc6587.txt - octet counting variant

	// 错误
	ErrIncompletePacket = errors.New("incomplete packet")
)

func (c *AutomaticCodec) GetParser(line []byte) parser.Parser {
	switch format := detect(line); format {
	case RFC3164:
		return rfc3164Parser
	case RFC5424:
		return rfc5424Parser
	default:
		return rfc3164Parser
	}
}

func (c *AutomaticCodec) Decode(conn gnet.Conn) ([]byte, parser.Parser, error) {
	buf, _ := conn.Next(-1)
	switch format := detect(buf); format {
	case RFC3164:
		return rfc3164Codec.Decode(conn)
	case RFC5424:
		return rfc5424Codec.Decode(conn)
	case RFC6587:
		return rfc6587Codec.Decode(conn)
	default:
		return rfc3164Codec.Decode(conn)
	}
}

/*
 * Will always fallback to rfc3164 (see section 4.3.3)
 */
func detect(data []byte) int {
	// all formats have a sapce somewhere
	if i := bytes.IndexByte(data, ' '); i > 0 {
		pLength := data[0:i]
		if _, err := strconv.Atoi(string(pLength)); err == nil {
			return RFC6587
		}
		// are we starting with <
		if data[0] != '<' {
			return RFC3164
		}
		// is there a close angle bracket before the ' '? there should be
		angle := bytes.IndexByte(data, '>')
		if (angle < 0) || (angle >= i) {
			return RFC3164
		}

		// if a single digit immediately follows the angle bracket, then a space
		// it is RFC5424, as RFC3164 must begin with a letter (month name)
		if (angle+2 == i) && (data[angle+1] >= '0') && (data[angle+1] <= '9') {
			return RFC5424
		} else {
			return RFC3164
		}
	}

	// fallback to rfc 3164 section 4.3.3
	return RFC3164
}
