package main

import (
	"DistributedFileSystems/p2p"
	"fmt"
	"log"
	"time"
)

func OnPeer(peer p2p.Peer) error {
	peer.Close()

	fmt.Println("doing some logic with the peer outside of TCPTransport")
	return nil
}

func main() {
	tcpTransportOpts := p2p.TCPTransportOps{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTcpTransport(tcpTransportOpts)
	fileServerOpts := FileServerOpts{
		ListenAddr:        ":3000",
		StorageRoot:       "3000_files",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
	}
	s := NewFileServer(fileServerOpts)

	go func() {
		time.Sleep(time.Second * 3)
		s.Stop()
	}()

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
