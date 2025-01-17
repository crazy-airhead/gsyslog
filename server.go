package gsyslog

import (
	"context"
	"github.com/crazy-airhead/gsyslog/codec"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"
	"strings"
)

var (
	rfc3164Codec   = &codec.RFC3164Codec{}   // RFC3164: http://www.ietf.org/rfc/rfc3164.txt
	rfc5424Codec   = &codec.RFC5424Codec{}   // RFC5424: http://www.ietf.org/rfc/rfc5424.txt
	rfc6587Codec   = &codec.RFC6587Codec{}   // RFC6587: http://www.ietf.org/rfc/rfc6587.txt - octet counting variant
	automaticCodec = &codec.AutomaticCodec{} // Automatically identify the codec
)

type Server struct {
	gnet.BuiltinEventEngine
	eng     gnet.Engine
	addr    string
	network string

	bufferSize int
	workerPool *goroutine.Pool

	codec   codec.Codec
	handler Handler
}

// NewServer returns a new Server
func NewServer() *Server {
	return &Server{
		handler:    NewDefaultHandler(),
		codec:      automaticCodec,
		workerPool: goroutine.Default(),
	}
}

// SetHandler Sets the handler, this handler with receive every syslog entry
func (s *Server) SetHandler(handler Handler) {
	s.handler = handler
}

// SetCodec Sets the syslog codec (RFC3164 or RFC5424 or RFC6587)
func (s *Server) SetCodec(f codec.Codec) {
	s.codec = f
}

// SetBufferSize Sets the maximum buffer size
func (s *Server) SetBufferSize(i int) {
	s.bufferSize = i
}

// SetBufferSize Sets the maximum buffer size
func (s *Server) SetAddr(addr string) {
	if strings.HasPrefix(addr, "udp://") {
		s.network = "udp"
		s.addr = addr
	} else if strings.HasPrefix(addr, "tcp://") {
		s.network = "tcp"
		s.addr = addr
	} else if strings.HasPrefix(addr, "unix://") {
		s.network = "tcp"
		s.addr = addr
	}
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

func (s *Server) OnTraffic(conn gnet.Conn) (action gnet.Action) {
	if s.network == "udp" {
		return s.handleUdp(conn)
	}

	if s.network == "tcp" {
		return s.handleTcp(conn)
	}

	if s.network == "unix" {
		return s.handleUdp(conn)
	}

	return gnet.None
}

func (s *Server) handleUdp(conn gnet.Conn) (action gnet.Action) {
	data, err := conn.Next(-1)
	if err != nil {
		logging.Errorf("syslog read buff, something wrong, error:%v", err)
		return gnet.None
	}

	client := conn.RemoteAddr().String()
	copyData := make([]byte, len(data))
	copy(copyData, data)
	_ = s.workerPool.Submit(func() {
		p := s.codec.GetParser(copyData)
		log, _ := p.Parse(copyData, client)
		s.handler.Handle(log)
	})

	return gnet.None
}

func (s *Server) handleTcp(conn gnet.Conn) (action gnet.Action) {
	for {
		data, p, err := s.codec.Decode(conn)
		if err != nil {
			break
		}

		client := conn.RemoteAddr().String()
		_ = s.workerPool.Submit(func() {
			log, _ := p.Parse(data, client)
			s.handler.Handle(log)
		})

		return gnet.None
	}

	if conn.InboundBuffered() > 0 {
		if err := conn.Wake(nil); err != nil { // wake up the connection manually to avoid missing the leftover data
			logging.Errorf("failed to wake up the connection, %v", err)
			return gnet.Close
		}
	}

	return gnet.None
}
