package main

import (
	"Project_6/p2p"
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
	for {
		select {
		case msg := <-fs.Opts.Transport.Consume():
			// Handle the incoming message
			fmt.Println(msg)
		// Process the message (e.g., store file, respond to requests)
		case <-fs.quitchan:
			return // Exit the loop when quitchan is closed
		}
	}
}

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
