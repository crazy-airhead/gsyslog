package gsyslog

import (
	"context"
	"github.com/crazy-airhead/gsyslog/format"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"
)

var (
	RFC3164   = &format.RFC3164{}   // RFC3164: http://www.ietf.org/rfc/rfc3164.txt
	RFC5424   = &format.RFC5424{}   // RFC5424: http://www.ietf.org/rfc/rfc5424.txt
	RFC6587   = &format.RFC6587{}   // RFC6587: http://www.ietf.org/rfc/rfc6587.txt - octet counting variant
	Automatic = &format.Automatic{} // Automatically identify the format
)

type Server struct {
	gnet.BuiltinEventEngine
	eng  gnet.Engine
	addr string

	bufferSize int
	workerPool *goroutine.Pool

	format  format.Format
	handler Handler
}

// NewServer returns a new Server
func NewServer() *Server {
	return &Server{
		handler:    NewDefaultHandler(),
		workerPool: goroutine.Default(),
	}
}

// SetHandler Sets the handler, this handler with receive every syslog entry
func (s *Server) SetHandler(handler Handler) {
	s.handler = handler
}

// SetFormat Sets the syslog format (RFC3164 or RFC5424 or RFC6587)
func (s *Server) SetFormat(f format.Format) {
	s.format = f
}

// SetBufferSize Sets the maximum buffer size
func (s *Server) SetBufferSize(i int) {
	s.bufferSize = i
}

// SetBufferSize Sets the maximum buffer size
func (s *Server) SetAddr(addr string) {
	s.addr = addr
}

func (s *Server) Boot() error {
	err := gnet.Run(s, s.addr,
		gnet.WithMulticore(true),
		gnet.WithSocketRecvBuffer(s.bufferSize))
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	_ = s.eng.Stop(context.Background())
	s.workerPool.Release()

	return nil
}

func (s *Server) OnBoot(eng gnet.Engine) gnet.Action {
	s.eng = eng

	logging.Infof("syslog server is listening on %s\n", s.addr)

	return gnet.None
}

func (s *Server) OnTraffic(c gnet.Conn) gnet.Action {
	data, err := c.Next(-1)
	if err != nil {
		logging.Errorf("syslog read loop, something wrong, error:%v", err)
		return gnet.None
	}

	task := func(client string, data []byte) {
		if sf := s.format.GetSplitFunc(); sf != nil {
			if _, token, err := sf(data, true); err == nil {
				s.parser(token, client, "")
			}
		} else {
			s.parser(data, client, "")
		}
	}

	client := c.RemoteAddr().String()
	//copyData := make([]byte, len(data))
	//copy(copyData, data)
	_ = s.workerPool.Submit(func() {
		task(client, data)
	})

	return gnet.None
}

func (s *Server) parser(line []byte, client string, tlsPeer string) {
	parser := s.format.GetParser(line)
	err := parser.Parse(client, tlsPeer)
	logParts := parser.Dump()

	s.handler.Handle(logParts, err)
}
