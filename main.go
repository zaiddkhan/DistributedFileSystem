package main

import (
	"DistributedFileSystems/p2p"
	"bytes"
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

func makeAServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(
		tcpTransportOpts,
	)

	fileServerOpts := FileServerOpts{
		EncKey:            newEncryptionKey(),
		StorageRoot:       strings.Replace(listenAddr, ":", "", -1) + "_network",
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
	s2 := makeAServer(":7000", "")
	s3 := makeAServer(":5000", ":3000", ":7000")

	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(500 * time.Millisecond)
	go func() { log.Fatal(s2.Start()) }()

	time.Sleep(2 * time.Second)

	go s3.Start()
	time.Sleep(2 * time.Second)

	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("picture_%d.png", i)
		data := bytes.NewReader([]byte("my big data file here!"))
		s3.StoreData(key, data)

		if err := s3.store.Delete(key); err != nil {
			log.Fatal(err)
		}

		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}

		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(b))
	}

}
