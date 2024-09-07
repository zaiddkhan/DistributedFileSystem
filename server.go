package main

import (
	"DistributedFileSystems/p2p"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts
	peerLock sync.Mutex
	peers    map[string]p2p.Peer
	store    *Store
	quitch   chan struct{}
}

type Payload struct {
	Key  string
	Data []byte
}

func (s *FileServer) broadcast(p *Payload) error {
	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(p)
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	buf := new(bytes.Buffer)

	tee := io.TeeReader(r, buf)

	if err := s.store.Write(key, tee); err != nil {
		return err
	}

	p := &Payload{Key: key, Data: buf.Bytes()}

	fmt.Println(buf.Bytes())
	return s.broadcast(p)
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		store:          NewStore(storeOpts),
		FileServerOpts: opts,
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("file server loop stopped")
		s.Transport.Close()
	}()
	for {
		select {
		case msg := <-s.Transport.Consume():
			log.Println("receive msg:", msg)
			var p Payload
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&p); err != nil {
				log.Println(err)
			}
		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) Store(key string, r io.Reader) error {
	return s.store.Write(key, r)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	log.Println("heheheh")
	s.peers[p.RemoteAddr().String()] = p
	log.Printf("connected with remote")
	return nil
}
func (s *FileServer) bootStrapMethod() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("transport error:", err)

			}
		}(addr)
	}
	return nil
}
func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootStrapMethod()

	s.loop()
	return nil
}
