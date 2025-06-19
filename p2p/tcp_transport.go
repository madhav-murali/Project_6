package p2p

import (
	"fmt"
	"net"
)

// Represents a peer in the P2P network. For the TCP implementation.
type TCPPeer struct {
	conn     net.Conn
	outbound bool   // true if outbound connection, false if inbound
	ID       string // Unique identifier for the peer, could be an IP address or a custom ID

}

type TCPTransportOpts struct {
	ListenAddress string
	ShakeHands    HandShakeFunc
	Decoder       Decoder // Decoder for decoding messages from the connection
}
type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener

	// peers map[net.Addr]Peer   // peer ID to connection
	// mu    sync.RWMutex
}

func NewTCPTransport(conf TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: conf,
		listener:         nil,
	}
}

func NewTCPPeer(conn net.Conn, outbound bool, id string) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
		ID:       id,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	}
	go t.StartAcceptLoop()
	return nil
}

func (t *TCPTransport) StartAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr().String())
		go t.handleConnection(conn)
	}

}

type Temp struct {
	// This is a placeholder for the message structure.
}

func (t *TCPTransport) handleConnection(conn net.Conn) {
	peer := NewTCPPeer(conn, false, conn.RemoteAddr().String())

	if err := t.ShakeHands(peer); err != nil {
		fmt.Printf("Handshake failed with peer %s: %v\n", peer.ID, err)
		conn.Close()
		return
	}

	msg := &Temp{}
	//could add a limit for error for a peer, if it fails to decode a message
	for {
		if err := t.Decoder.Decode(conn, msg); err != nil {
			fmt.Printf("Error decoding message from peer %s: %v\n", peer.ID, err)
			conn.Close()
			return
		}
	}

	// Handle the connection (e.g., read/write data)
	// This is a placeholder for actual handling logic
	// You can implement your own protocol here
}
