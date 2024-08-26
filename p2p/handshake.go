package p2p

type Handshaker interface {
	Handshake() error
}

// handshake func is
type HandshakeFunc func(any) error

type DefaultHandshaker struct {
}

func NOPHandshakeFunc(any) error {
	return nil
}
