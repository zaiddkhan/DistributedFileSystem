package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTcpTransport(t *testing.T) {
	opts := TCPTransportOps{
		ListenAddr:    ":3000",
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTcpTransport(opts)
	assert.Equal(t, tr.listenAddr, ":3000")

	assert.Nil(t, tr.ListenAndAccept())
	select {}

}
