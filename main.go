package main

import (
	"DistributedFileSystems/p2p"
	"fmt"
)

func OnPeer(peer p2p.Peer) error {
	peer.Close()

	fmt.Println("doing some logic with the peer outside of TCPTransport")
	return nil
}

func main() {
	tcpOpts := p2p.TCPTransportOps{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tr := p2p.NewTcpTransport(tcpOpts)

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("$+v\n", msg)
		}
	}()
	if err := tr.ListenAndAccept(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("listen and accept ok")

	select {}

}
