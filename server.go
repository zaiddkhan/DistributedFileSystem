package main

import (
	"DistributedFileSystems/p2p"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
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
type MessageGetFile struct {
	Key string
}

type Message struct {
	From    string
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

func (s *FileServer) stream(msg *Message) error {
	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) broadcast(msg *Message) error {
	msgBuf := new(bytes.Buffer)

	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}

	}
	return nil
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		_, r, err := s.store.Read(key)
		return r, err
	}
	fmt.Printf("don't have file locally")
	msg := Message{
		Payload: MessageGetFile{
			key,
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}
	time.Sleep(time.Millisecond * 3)
	for _, peer := range s.peers {
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		n, err := s.store.Write(key, io.LimitReader(peer, fileSize))
		if err == nil {
			return nil, err
		}

		fmt.Println(n)
		peer.CloseStream()

	}
	_, r, err := s.store.Read(key)
	return r, err
}
func (s *FileServer) StoreData(key string, r io.Reader) error {

	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)

	size, err := s.store.Write(key, tee)

	if err != nil {
		return err
	}
	msg := &Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size + 16,
		},
	}

	if err := s.broadcast(msg); err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 5)

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingStream})
		n, err := io.Copy(peer, fileBuffer)
		if err != nil {
			return err
		}
		fmt.Println(n)

	}

	return nil
}

func (s *FileServer) handleMessage(from string, p *Message) error {
	switch v := p.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		fmt.Println("hehehe")
		if err := s.handleMessageGetFile(from, &v); err != nil {
			return err
		}

	}

	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg *MessageGetFile) error {
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("cannot read file (%s) from disk", msg.Key)
	}
	size, r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("cannot read file (%s) from disk", msg.Key)
	}

	rc, ok := r.(io.ReadCloser)
	if ok {
		fmt.Println("closing read closer")
		defer rc.Close()
	}
	peer.Send([]byte{p2p.IncomingStream})
	var fileSize int64 = size
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}
	fmt.Printf("written %d bytes over the netwrok to %s\n", n, from)
	return nil

}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not found", from)
	}
	if _, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size)); err != nil {
		return err
	}
	peer.CloseStream()

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
		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) Store(key string, r io.Reader) (int64, error) {
	return s.store.Write(key, r)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
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
	gob.Register(MessageGetFile{})
}
