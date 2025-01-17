package gsyslog

import (
	"testing"
)

func Test_udp_server(t *testing.T) {
	server := NewServer()
	server.SetCodec(rfc3164Codec)
	server.SetAddr("udp://0.0.0.0:514")
	defer func(server *Server) {
		_ = server.Stop()
	}(server)

	_ = server.Boot()
}

func Test_tcp_server(t *testing.T) {
	server := NewServer()
	server.SetCodec(rfc3164Codec)
	server.SetAddr("tcp://0.0.0.0:514")
	defer func(server *Server) {
		_ = server.Stop()
	}(server)

	_ = server.Boot()
}
