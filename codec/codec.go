package codec

import (
	"errors"
	"github.com/crazy-airhead/gsyslog/parser"
	"github.com/crazy-airhead/gsyslog/parser/rfc3164"
	"github.com/crazy-airhead/gsyslog/parser/rfc5424"
	"github.com/panjf2000/gnet/v2"
)

type Codec interface {
	Decode(conn gnet.Conn) ([]byte, error)
	GetParser([]byte) parser.Parser
}

var (
	rfc3164Parser = rfc3164.NewParser() // RFC3164: http://www.ietf.org/rfc/rfc3164.txt
	rfc5424Parser = rfc5424.NewParser() // RFC5424: http://www.ietf.org/rfc/rfc5424.txt
)

var (
	ErrIncompletePacket = errors.New("incomplete packet")
)
