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

type Message struct {
	From    string
	Payload any
}

type MessageStoreFile struct {
	Key string
}

func (s *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) StoreData(key string, r io.Reader) error {

	buf := new(bytes.Buffer)
	msg := &Message{
		Payload: MessageStoreFile{
			Key: key,
		},
	}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {

		}

	}
	//time.Sleep(time.Second * 3)
	//payload := []byte("LARGE FILE")
	//
	//for _, peer := range s.peers {
	//	if err := peer.Send(payload); err != nil {
	//
	//	}
	//
	//}

	return nil
	//buf := new(bytes.Buffer)
	//
	//tee := io.TeeReader(r, buf)
	//
	//if err := s.store.Write(key, tee); err != nil {
	//	return err
	//}
	//
	//p := &DataMessage{Key: key, Data: buf.Bytes()}
	//
	//fmt.Println(buf.Bytes())
	//return s.broadcast(&Message{
	//	From:    "todo",
	//	Payload: p,
	//})
}

func (s *FileServer) handleMessage(from string, p *Message) error {
	switch v := p.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	}
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not found", from)
	}
	if err := s.store.Write(msg.Key, peer); err != nil {
		return err
	}
	peer.(*p2p.TCPPeer).Wg.Done()

	return nil
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
		case rpc := <-s.Transport.Consume():
			log.Println("receive msg:", rpc)
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println(err)
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println(err)
			}
			//fmt.Printf("%+v\n", msg.Payload)
			//peer, ok := s.peers[rpc.From]
			//if !ok {
			//	panic("no peer")
			//}
			//b := make([]byte, 1024)
			//if _, err := peer.Read(b); err != nil {
			//	log.Println(err)
			//}
			//fmt.Println(peer)
			//
			//fmt.Printf("recv %s", string(msg.Payload.([]byte)))
			//if err := s.handleMessage(&p); err != nil {
			//	log.Println(err)
			//}
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

func init() {
	gob.Register(MessageStoreFile{})
}
