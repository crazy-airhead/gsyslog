package codec

import (
	"github.com/crazy-airhead/gsyslog/parser"
	"github.com/panjf2000/gnet/v2"
)

type Codec interface {
	Decode(conn gnet.Conn) ([]byte, error)
	GetParser([]byte) parser.Parser
}
