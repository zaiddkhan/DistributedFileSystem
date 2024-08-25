package p2p

type Handshaker interface {
	Handshake() error
}

type HandshakeFunc func(any) error

type DefaultHandshaker struct {
}
