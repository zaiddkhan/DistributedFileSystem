package main

import (
	"DistributedFileSystems/p2p"
	"fmt"
	"log"
)

func OnPeer(peer p2p.Peer) error {
	peer.Close()

	fmt.Println("doing some logic with the peer outside of TCPTransport")
	return nil
}

func makeAServer(listenAddr string, root string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOps{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTcpTransport(
		tcpTransportOpts,
	)

	fileServerOtpts := FileServerOpts{
		StorageRoot:       listenAddr + "3000_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}
	return NewFileServer(fileServerOtpts)
}

func main() {
	s1 := makeAServer(":3000", "")
	s2 := makeAServer(":4000", ":3000")
	go func() {
		log.Fatal(s1.Start())
	}()
	s2.Start()
}
