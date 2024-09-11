package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

//tcp peer represents the remote node over a TCP established connection

type TCPPeer struct {
	//conn is the underlying connection of the peer
	net.Conn

	//if we dial and retrieve a connection => outbound == true
	//if we accept and retrieve a connection => outbound == false
	outbound bool
	Wg       *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		Wg:       &sync.WaitGroup{},
	}
}

type TCPTransportOps struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOps
	listener net.Listener
	rpcch    chan RPC
}

func NewTcpTransport(opts TCPTransportOps) *TCPTransport {
	return &TCPTransport{
		TCPTransportOps: opts,
		rpcch:           make(chan RPC, 1024),
	}
}

func (t *TCPTransport) Addr() string {
	return t.ListenAddr
}

// Implements the transport interface, which will return a read only channel
// for reading the incoming messages recevied from another peer.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()
	log.Printf("TCP transport listening on", t.ListenAddr)

	return nil
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

// Dial implements the Transport interface.
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true)
	return nil

}
func (t *TCPTransport) startAcceptLoop() {

	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {

			fmt.Println("accept err:", err)
		}
		fmt.Printf("new incoming connection %v\n", conn)

		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	defer func() {

		fmt.Printf("dropping peer connection %s", err)
		conn.Close()

	}()
	peer := NewTCPPeer(conn, true)

	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Println("tcp handshake err:", err)
		return
	}

	if t.OnPeer != nil {
		if err := t.OnPeer(peer); err != nil {
			return
		}
	}
	//read loop

	for {
		rpc := RPC{}
		err := t.Decoder.Decode(conn, &rpc)

		if err != nil {

			fmt.Printf("TCP read error: %s\n", err)
			return
		}
		if rpc.Stream {
			peer.Wg.Add(1)
			fmt.Println("waitinggggg")
			peer.Wg.Wait()
			continue
		}
		rpc.From = conn.RemoteAddr().String()
		t.rpcch <- rpc

	}
}
