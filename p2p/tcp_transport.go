package p2p

import (
	"fmt"
	"net"
	"sync"
)

//tcp peer represents the remote node over a TCP established connection

type TCPPeer struct {
	//conn is the underlying connection of the peer
	conn net.Conn

	//if we dial and retrieve a connection => outbound == true
	//if we accept and retrieve a connection => outbound == false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransport struct {
	listenAddr    string
	listener      net.Listener
	handshakeFunc HandshakeFunc

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTcpTransport(listenerAddr string) *TCPTransport {
	return &TCPTransport{
		handshakeFunc: func(any) error { return nil },
		listenAddr:    listenerAddr,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}
	go t.startAcceptLoop()
	return nil
}

func (t *TCPTransport) startAcceptLoop() error {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Println("accept err:", err)
		}

		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	fmt.Printf("new incoming connection %+v\n", peer)
}
