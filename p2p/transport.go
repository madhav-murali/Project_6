package p2p

//The users
type Peer interface {
}

//The layer for communication TCP,UDP,WebSockets, etc.
type Transport interface {
	ListenAndAccept() error
}
