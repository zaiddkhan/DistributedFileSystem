package main

import (
	"DistributedFileSystems/p2p"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

func OnPeer(peer p2p.Peer) error {
	peer.Close()

	fmt.Println("doing some logic with the peer outside of TCPTransport")
	return nil
}

func makeAServer(listenAddr string, root string, nodes ...string) *FileServer {
	addrs := strings.ReplaceAll(listenAddr, ":", "")
	tcpTransportOpts := p2p.TCPTransportOps{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
	}
	tcpTransport := p2p.NewTcpTransport(
		tcpTransportOpts,
	)

	fileServerOpts := FileServerOpts{
		StorageRoot:       addrs + "_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}
	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}

func main() {
	s1 := makeAServer(":3000", "")
	s2 := makeAServer(":4000", ":3000")
	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(2 * time.Second)
	go s2.Start()
	time.Sleep(2 * time.Second)
	r, err := s2.Get("myprivdata")
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)

	}
	//data := bytes.NewReader([]byte("big data found"))
	//s2.StoreData("myprivdata", data)
	fmt.Println(string(b))
	select {}
}
