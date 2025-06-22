package p2p

//The users
type Peer interface {
	Close() error
	ID() string // Unique identifier for the peer, could be an IP address or a custom ID
}

//The layer for communication TCP,UDP,WebSockets, etc.
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC // Channel to receive messages from peers
}
