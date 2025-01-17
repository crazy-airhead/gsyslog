package codec

import (
	"github.com/crazy-airhead/gsyslog/parser"
	"github.com/panjf2000/gnet/v2"
)

type RFC3164Codec struct{}

func (f *RFC3164Codec) GetParser(data []byte) parser.Parser {
	return rfc3164Parser
}

func (f *RFC3164Codec) Decode(conn gnet.Conn) ([]byte, parser.Parser, error) {
	buf, _ := conn.Next(-1)

	length := len(buf)
	body := make([]byte, length)
	copy(body, buf)

	_, _ = conn.Discard(length)

	return body, rfc3164Parser, nil
}
