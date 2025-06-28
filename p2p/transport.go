package p2p

import (
	"net"
)

// The users
type Peer interface {
	net.Conn
	Send(msg *RPC) error // Send a message to the peer // WaitGroup to wait for the peer to finish processing
	//*sync.WaitGroup      // WaitGroup to wait for the peer to finish processing
	// Close() error
	// RemoteAddr() net.Addr // Unique identifier for the peer, could be an IP address or a custom ID
}

// The layer for communication TCP,UDP,WebSockets, etc.
type Transport interface {
	Dial(addr string) error // Dial a peer by address
	ListenAndAccept() error
	Consume() <-chan RPC // Channel to receive messages from peers
	Close() error        // Close the transport to free resources
	//ListenAddr() string  // Get the address of the transport
}
