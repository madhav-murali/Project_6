package p2p

import "net"

type RPC struct {
	From    net.Addr // The address of the peer sending the message
	To      net.Addr // The address of the peer receiving the message
	Payload []byte   // The actual data being sent in the message
}
