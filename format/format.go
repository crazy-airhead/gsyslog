package format

import (
	"bufio"
	"github.com/crazy-airhead/gsyslog/parser"
)

type Format interface {
	GetParser([]byte) parser.LogParser
	GetSplitFunc() bufio.SplitFunc
}

type parserWrapper struct {
	parser.LogParser
}

func (w *parserWrapper) Dump() *parser.Log {
	return w.LogParser.Dump()
}
