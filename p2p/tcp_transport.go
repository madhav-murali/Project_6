package p2p

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"sync"
)

// Represents a peer in the P2P network. For the TCP implementation.
type TCPPeer struct {
	net.Conn
	outbound bool            // true if outbound connection, false if inbound
	Wg       *sync.WaitGroup // Embed WaitGroup to satisfy interfaces that require it
	//ID       string // Unique identifier for the peer, could be an IP address or a custom ID
}

// Returns the address of the transport.
// This is used to get the address of the transport for sending messages.
func (t *TCPTransport) ListenAddr() string {
	return t.listener.Addr().String() // Return the address of the listener
}

// RemoteAddr returns the unique identifier for the peer to satisfy the Peer interface.
func (p *TCPPeer) RemoteAddr() net.Addr {
	return p.Conn.RemoteAddr()
}

// Send sends a message to the peer over the TCP connection.
// It encodes the message and writes it to the connection.
// It returns an error if the connection is nil or if there is an error sending the message
func (p *TCPPeer) Send(msg *RPC) error {
	if p.Conn == nil {
		return errors.New("connection is nil, cannot send message")
	}
	//fmt.Printf("Sending message to %s: key=%s, payload=%s\n", p.Conn.RemoteAddr(), msg.From, string(msg.Payload))

	//Encode the message and send it over the connection.
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(msg); err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}
	data := buf.Bytes()

	lengthPrefix := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthPrefix, uint32(len(data)))

	if _, err := p.Conn.Write(lengthPrefix); err != nil {
		return fmt.Errorf("failed to write length prefix: %w", err)
	}
	if _, err := p.Conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
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

func NewTCPPeer(conn net.Conn, outbound bool, addr net.Addr) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		Wg:       &sync.WaitGroup{}, // Wg is zero value initialized automatically
		//ID:       id,
	}
}

// Close closes the connection to the peer.
// It is important to close the connection when done to free up resources.
func (p *TCPPeer) Close() error {
	if p.Conn != nil {
		return p.Conn.Close()
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

	fmt.Printf("TCP transport listening on %s\n", t.ListenAddress)

	return nil
}

func (t *TCPTransport) StartAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			fmt.Println("Listener closed, stopping accept loop")
			return // Exit the loop if the listener is closed
		}
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr())
		go t.handleConnection(conn, false) // Handle the connection in a separate goroutine
	}

}

// Close closes the transport and frees resources.
// It closes the listener and the rpcchan channel to stop receiving messages.
// It is important to close the transport when done to free up resources.
func (t *TCPTransport) Close() error {
	return t.listener.Close() // Close the listener to stop accepting new connections

}

// Dial connects to a peer by address.
// It creates a new connection to the peer and starts handling it in a separate goroutine.
// It returns an error if the connection fails.
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("error dialing peer %s: %w", addr, err)

	}
	go t.handleConnection(conn, true) // Handle the connection in a separate goroutine
	return nil
}

// handleConnection handles the incoming connection from a peer.
// It performs the handshake, decodes messages, and sends them to the rpcchan channel.
func (t *TCPTransport) handleConnection(conn net.Conn, outbound bool) {
	if outbound {
		fmt.Printf("Outbound connection established to peer: %s\n", conn.RemoteAddr())
	} else {
		fmt.Printf("Inbound connection established from peer: %s\n", conn.RemoteAddr())
	}
	var err error
	defer func() {
		fmt.Printf("Closing connection to peer: %s\n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, false, conn.RemoteAddr())
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

	//rpc := RPC{}
	//could add a limit for error for a peer, if it fails to decode a message
	for {
		rpc := RPC{}
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}

		rpc.From = conn.RemoteAddr().String()

		if rpc.Stream {
			peer.Wg.Add(1)
			fmt.Printf("[%s] incoming stream, waiting...\n", conn.RemoteAddr())
			t.rpcchan <- rpc
			peer.Wg.Wait()
			fmt.Printf("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
		}

	}

}
