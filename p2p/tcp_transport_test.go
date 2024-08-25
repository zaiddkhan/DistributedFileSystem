package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTcpTransport(t *testing.T) {

	listenAdd := ":9000"
	tr := NewTcpTransport(listenAdd)
	assert.Equal(t, listenAdd, tr.listenAddr)

	assert.Nil(t, tr.ListenAndAccept())
	select {}

}
