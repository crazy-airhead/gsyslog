package codec

import (
	"bytes"
	"github.com/crazy-airhead/gsyslog/parser"
	"github.com/panjf2000/gnet/v2"
)

type RFC5424Codec struct{}

func (f *RFC5424Codec) GetParser(data []byte) parser.Parser {
	return rfc5424Parser
}

func (f *RFC5424Codec) Decode(conn gnet.Conn) ([]byte, error) {
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

func rfc5424ScannerSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	// return all of the data without splitting
	return len(data), data, nil
}
