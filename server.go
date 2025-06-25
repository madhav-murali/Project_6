package main

import (
	"Project_6/p2p"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
)

type FileServerOpts struct {
	ListenAddr        string
	StorageRoot       string
	PathTransformFunc PathTransformFunc // Function to transform keys into paths
	Transport         p2p.Transport
	BootStrapNodes    []string // List of bootstrap nodes to connect to
	//TCPTransportOpts  p2p.TCPTransportOpts // Options for TCP transport
	// Other options can be added as needed
}

type FileServer struct {
	Opts     FileServerOpts
	PeerLock sync.Mutex          // Mutex to protect access to peers
	Peers    map[string]p2p.Peer // Map of connected peers
	Store    *Store
	quitchan chan struct{} // Channel to signal shutdown
}

func NewFileServer(opts FileServerOpts) *FileServer {
	if opts.StorageRoot == "" {
		opts.StorageRoot = defaultRootFolderName // Set default root folder if not provided
	}
	// Create a new store with the provided options
	store := NewStore(StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	})
	return &FileServer{
		Opts:     opts,
		Peers:    make(map[string]p2p.Peer), // Initialize the map of peers
		Store:    store,
		quitchan: make(chan struct{}), // Initialize the quit channel
	}
}

type Message struct {
	Payload any
}

// Broadcast sends a message to all connected peers.
func (fs *FileServer) Broadcast(msg *Message) error {
	// if msg == nil {
	// 	return fmt.Errorf("message cannot be nil")
	// }
	// fs.PeerLock.Lock()
	// defer fs.PeerLock.Unlock()
	// log.Printf("Broadcasting to %d peers", len(fs.Peers))

	// for _, peer := range fs.Peers {
	// 	var buf bytes.Buffer
	// 	if err := gob.NewEncoder(&buf).Encode(msg); err != nil {
	// 		log.Printf("Failed to encode message for peer %s: %v", peer.RemoteAddr(), err)
	// 		continue
	// 	}
	// 	if _, err := peer.Write(buf.Bytes()); err != nil {
	// 		log.Printf("Failed to send message to peer %s: %v", peer.RemoteAddr(), err)
	// 	} else {
	// 		log.Printf("Message sent to peer %s: %s", peer.RemoteAddr(), msg.Key)
	// 	}
	// }
	// return nil

	peers := []io.Writer{}
	for _, peer := range fs.Peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (fs *FileServer) StoreData(key string, r io.Reader) error {
	//1.store the data in the disk
	// 2. broadcast the message to all peers
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		return err
	}

	//payload := []byte(buf.Bytes()) // Convert the buffer to a byte slice

	msg := p2p.RPC{
		Key:     key,
		Payload: buf.Bytes(), // Set the payload to the byte slice
	}

	// if err := gob.NewEncoder(buf).Encode(msg); err != nil {
	// 	return fmt.Errorf("failed to encode message: %v", err)
	// }
	// Store the file in the store
	for _, peer := range fs.Peers {
		if err := peer.Send(&msg); err != nil {
			log.Printf("Failed to send message to peer %s: %v", peer.RemoteAddr(), err)
			continue
		}
		log.Printf("Message sent to peer %s: %s", peer.RemoteAddr(), key)
	}
	return nil
}

// Stop gracefully stops the file server.
func (fs *FileServer) Stop() {
	// close(fs.quitchan) // Signal the loop to stop
	// if fs.Opts.Transport != nil {
	// 	fs.Opts.Transport.Close() // Close the transport if it implements Close
	// }
	close(fs.quitchan) // Signal the loop to stop
}

// AddPeer adds a new peer to the file server.
// It locks the PeerLock mutex to ensure thread-safe access to the Peers map.
func (fs *FileServer) AddPeer(peer p2p.Peer) error {
	fs.PeerLock.Lock()
	defer fs.PeerLock.Unlock()
	fs.Peers[peer.RemoteAddr().String()] = peer // Add the peer to the map of peers
	log.Printf("Peer added to connection : %s", peer.RemoteAddr().String())
	return nil
}

// loop handles incoming messages and processes them.
// It runs in a separate goroutine to allow asynchronous message handling.
// The loop will continue until the quitchan is closed.
func (fs *FileServer) loop() {
	defer func() {
		log.Println("File server loop stopped")
		fs.Opts.Transport.Close() // Ensure transport is closed when the loop stops
	}()
	fmt.Println("File server loop started")
	for {
		select {
		case rpc := <-fs.Opts.Transport.Consume():
			switch payload := rpc.Payload.(type) {
			case []byte:
				log.Printf("Received raw key/string: %s", string(payload))
			default:
				log.Printf("Received unknown message type: %T", payload)
				// You can handle other message types here if needed
			}
		case <-fs.quitchan:
			return // Exit the loop when quitchan is closed
		}
	}
}

// func (fs *FileServer) HandleMessage(msg *Message) error {

// 	switch payload := msg.Payload.(type) {
// 	case *DataMessage:
// 		log.Printf("Received data message with key: %+v", payload)

// 	}
// 	return nil
// }

// bootstrapNetwork connects to the bootstrap nodes specified in the FileServer options.
// It runs in a separate goroutine for each bootstrap node to allow concurrent connections.
func (fs *FileServer) bootstrapNetwork() error {

	for _, addr := range fs.Opts.BootStrapNodes {
		if addr == "" {
			continue // Skip empty addresses
		}
		go func(addr string) {
			if err := fs.Opts.Transport.Dial(addr); err != nil {
				log.Printf("Failed to connect to bootstrap node %s: %v", addr, err)
			} else {
				log.Printf("Connected to bootstrap node %s", addr)
			}
		}(addr)

	}
	return nil
}

func (fs *FileServer) Start() error {
	if err := fs.Opts.Transport.ListenAndAccept(); err != nil {
		return err
	}

	fs.bootstrapNetwork() // Connect to bootstrap nodes

	fs.loop() // Start the loop to handle incoming messages
	return nil
}

func (fs *FileServer) StoreFile(key string, r io.Reader) error {
	// Store the file in the store
	return fs.Store.Write(key, r)
}
