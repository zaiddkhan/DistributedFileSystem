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

type TCPTransportOps struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
}

type TCPTransport struct {
	TCPTransportOps
	listenAddr string
	listener   net.Listener
	rpcch      chan RPC

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTcpTransport(opts TCPTransportOps) *TCPTransport {
	return &TCPTransport{
		TCPTransportOps: opts,
		rpcch:           make(chan RPC),
	}
}

// implements the transport interface, which will return a read only channel
// for reading the incoming messages recevied from another peer.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
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
		fmt.Printf("new incoming connection %v\n", conn)

		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Println("tcp handshake err:", err)
		return
	}

	//read loop
	rpc := RPC{}
	for {
		if err := t.Decoder.Decode(conn, &rpc); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}
		rpc.from = conn.RemoteAddr()
		t.rpcch <- rpc
		fmt.Printf("TCP msg: %v\n", rpc)
	}
}
