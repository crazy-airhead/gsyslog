package codec

import (
	"bytes"
	"github.com/crazy-airhead/gsyslog/parser"
	"github.com/panjf2000/gnet/v2"
)

type RFC3164Codec struct{}

func (f *RFC3164Codec) GetParser(data []byte) parser.Parser {
	return rfc3164Parser
}

func (f *RFC3164Codec) Decode(conn gnet.Conn) ([]byte, error) {
	buf, _ := conn.Next(-1)
	idx := bytes.IndexByte(buf, '\n')
	if idx == -1 {
		// 如果没有找到换行符，说明数据不完整，等待更多数据
		return nil, ErrIncompletePacket
	}

	body := make([]byte, idx)
	copy(body, buf[:idx])

	// 往前移动
	_, _ = conn.Discard(idx)

	return body, nil
}
