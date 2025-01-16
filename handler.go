package gsyslog

import (
	"github.com/crazy-airhead/gsyslog/parser"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"sync/atomic"
)

// Handler The handler receive every syslog entry at Handle method
type Handler interface {
	Handle(log *parser.Log)
}

type DefaultHandler struct {
	counter *int64
}

func NewDefaultHandler() *DefaultHandler {
	return &DefaultHandler{
		counter: new(int64),
	}
}

// Handle entry receiver
func (h *DefaultHandler) Handle(log *parser.Log) {
	atomic.AddInt64(h.counter, 1)

	logging.Infof("number %d, header:%v, body:%v,", *h.counter, log.Header, log.GetString(parser.LogBody))
}
