package p2p

import "net"

// RPC holds any arbitrary data that is being sent over
// each transport between two nodes in the network
type RPC struct {
	from    net.Addr
	Payload []byte
}
