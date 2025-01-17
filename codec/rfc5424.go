package codec

import (
	"github.com/crazy-airhead/gsyslog/parser"
	"github.com/panjf2000/gnet/v2"
)

type RFC5424Codec struct{}

func (f *RFC5424Codec) GetParser(data []byte) parser.Parser {
	return rfc5424Parser
}

func (f *RFC5424Codec) Decode(conn gnet.Conn) ([]byte, parser.Parser, error) {
	buf, _ := conn.Next(-1)

	length := len(buf)
	body := make([]byte, length)
	copy(body, buf)

	_, _ = conn.Discard(length)

	return body, rfc5424Parser, nil
}
