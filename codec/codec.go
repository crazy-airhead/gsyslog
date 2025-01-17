package codec

import (
	"github.com/crazy-airhead/gsyslog/parser"
	"github.com/panjf2000/gnet/v2"
)

type Codec interface {
	// Decode 用于 TCP 拆包
	Decode(conn gnet.Conn) ([]byte, parser.Parser, error)
	// GetParser 获取解析器，用于格式解析
	GetParser([]byte) parser.Parser
}
