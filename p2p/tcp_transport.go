package p2p

import (
	"fmt"
	"net"
)

// Represents a peer in the P2P network. For the TCP implementation.
type TCPPeer struct {
	conn     net.Conn
	outbound bool // true if outbound connection, false if inbound
	//ID       string // Unique identifier for the peer, could be an IP address or a custom ID
}

// ID returns the unique identifier for the peer to satisfy the Peer interface.
func (p *TCPPeer) ID() string {
	return p.conn.RemoteAddr().String()
}

type TCPTransportOpts struct {
	ListenAddress string
	ShakeHands    HandShakeFunc
	Decoder       Decoder          // Decoder for decoding messages from the connection
	OnPeerConnect func(Peer) error // Callback function when a peer connects
}
type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcchan  chan RPC // Channel to receive messages from peers
	// peers map[net.Addr]Peer   // peer ID to connection
	// mu    sync.RWMutex
}

func NewTCPTransport(conf TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: conf,
		listener:         nil,
		rpcchan:          make(chan RPC, 100), // Buffered channel for receiving messages
	}
}

// Consume implements the Transport interface for TCPTransport. Which will return a channel for read only.
// This channel will be used to receive messages from peers.
// The channel is to be buffered to allow for asynchronous message handling.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcchan
}

func NewTCPPeer(conn net.Conn, outbound bool, id string) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
		//ID:       id,
	}
}

// Close closes the connection to the peer.
// It is important to close the connection when done to free up resources.
func (p *TCPPeer) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
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

// handleConnection handles the incoming connection from a peer.
// It performs the handshake, decodes messages, and sends them to the rpcchan channel.
func (t *TCPTransport) handleConnection(conn net.Conn) {
	var err error
	defer func() {
		fmt.Printf("Closing connection to peer: %s\n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, false, conn.RemoteAddr().String())
	if err := t.ShakeHands(peer); err != nil {
		fmt.Printf("Handshake failed with peer : %v\n", err)
		conn.Close()
		return
	}

	// Invoke the OnPeerConnect callback if it is set.
	// This allows for custom actions to be performed when a peer connects.
	if t.OnPeerConnect != nil {
		if err := t.OnPeerConnect(peer); err != nil {
			fmt.Printf("Error in OnPeerConnect callback: %v\n", err)
			conn.Close()
			return
		}
	}

	rpc := RPC{}
	spamProtec := 0
	//could add a limit for error for a peer, if it fails to decode a message
	for {
		if err := t.Decoder.Decode(conn, &rpc); err != nil {
			fmt.Printf("Error decoding message from peer: %v\n", err)
			conn.Close()
			spamProtec++
			if spamProtec > 5 {
				return
			} // Limit the number of errors before closing the connection
			continue
		}
		rpc.From = conn.RemoteAddr()
		t.rpcchan <- rpc // Send the decoded message to the channel
		//fmt.Printf("Received message from peer : %+v\n", string(rpc.Payload))
	}

	// Handle the connection (e.g., read/write data)
	// This is a placeholder for actual handling logic
	// You can implement your own protocol here
}
