package main

import (
	"DistributedFileSystems/p2p"
	"fmt"
)

func main() {
	tcpOpts := p2p.TCPTransportOps{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tr := p2p.NewTcpTransport(tcpOpts)
	if err := tr.ListenAndAccept(); err != nil {
		fmt.Println(err)
	}

	select {}
}
