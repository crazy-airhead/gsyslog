package gsyslog

import (
	"testing"
)

func Test_server(t *testing.T) {
	server := NewServer()
	server.SetCodec(RFC3164Codec)
	server.SetAddr("udp://0.0.0.0:514")
	defer func(server *Server) {
		_ = server.Stop()
	}(server)

	_ = server.Boot()
}
