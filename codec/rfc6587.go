package codec

import (
	"encoding/binary"
	"github.com/crazy-airhead/gsyslog/parser"
	"github.com/crazy-airhead/gsyslog/parser/rfc5424"
	"github.com/panjf2000/gnet/v2"
)

type RFC6587Codec struct{}

func (f *RFC6587Codec) GetParser(data []byte) parser.Parser {
	return rfc5424.NewParser()
}

func (f *RFC6587Codec) Decode(conn gnet.Conn) ([]byte, error) {
	if conn.InboundBuffered() < 4 {
		// 如果没有找到换行符，说明数据不完整，等待更多数据
		return nil, ErrIncompletePacket
	}

	lenBuf, _ := conn.Peek(4)
	length := binary.BigEndian.Uint32(lenBuf)

	_, _ = conn.Discard(4)

	buf, _ := conn.Next(int(length))
	body := make([]byte, length)
	copy(body, buf)

	_, _ = conn.Discard(int(length))

	return nil, nil
}
