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

type MessageStoreFile struct {
	Key  string // Key to identify the file
	Size int64  // Size of the file, if needed

}

type MessageGetFile struct {
	Key string // Key to identify the file

}

// Broadcast sends a message to all connected peers.
func (fs *FileServer) Broadcast(msg *Message) error {

	peers := []io.Writer{}
	for _, peer := range fs.Peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (fs *FileServer) StoreData(key string, r io.Reader) error {

	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf) // Create a TeeReader to read from r and write to buf

	if _, err := fs.Store.writeStream(key, tee); err != nil {
		return fmt.Errorf("failed to store data: %v", err)
	}

	//buf := new(bytes.Buffer)
	msg := &Message{
		Payload: &MessageStoreFile{
			Key:  key, // Set the key for the file
			Size: 16,
		},
	}

	rpc1 := p2p.RPC{
		From:    key,
		Payload: msg,  // Set the payload to the byte slice
		Stream:  true, // Indicate that this is a stream
	}

	for _, peer := range fs.Peers {
		if err := peer.Send(&rpc1); err != nil {
			log.Printf("failed to send message to peer %s: %v", peer.RemoteAddr(), err)
			continue
		}
		n, err := io.Copy(peer, r)
		if err != nil {
			return err
		}
		fmt.Println("Written and received bytes  ; ", n)
	}

	//payload := []byte("Hello, this is a test message") // Example payload
	// rpc := p2p.RPC{
	// 	From:    key,
	// 	Payload: payload, // Set the payload to the byte slice
	// }s

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
			//var msg Message
			msg, ok := rpc.Payload.(*Message)
			if !ok {
				log.Println("received payload is not of type *Message")
				continue // Skip processing if the payload is not of type *Message
			}
			if err := fs.handleMessage(rpc.From, msg); err != nil {
				log.Println("handle message error: ", err)
				return
			}

		case <-fs.quitchan:
			return // Exit the loop when quitchan is closed
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
		// case MessageGetFile:
		// 	return s.handleMessageGetFile(from, v)
	}

	return nil
}

// func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
// 	if !s.Store.HasKey(msg.ID, msg.Key) {
// 		return fmt.Errorf("[%s] need to serve file (%s) but it does not exist on disk", s.Transport.Addr(), msg.Key)
// 	}

// 	fmt.Printf("[%s] serving file (%s) over the network\n", s.Transport.Addr(), msg.Key)

// 	fileSize, r, err := s.store.Read(msg.ID, msg.Key)
// 	if err != nil {
// 		return err
// 	}

// 	if rc, ok := r.(io.ReadCloser); ok {
// 		fmt.Println("closing readCloser")
// 		defer rc.Close()
// 	}

// 	peer, ok := s.Peers[from]
// 	if !ok {
// 		return fmt.Errorf("peer %s not in map", from)
// 	}

// 	// First send the "incomingStream" byte to the peer and then we can send
// 	// the file size as an int64.
// 	peer.Send([]byte{p2p.IncomingStream})
// 	binary.Write(peer, binary.LittleEndian, fileSize)
// 	n, err := io.Copy(peer, r)
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Printf("[%s] written (%d) bytes over the network to %s\n", s.Transport.Addr(), n, from)

// 	return nil
// }

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.Peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}
	fmt.Printf("[%s] received file store request for key (%s) from peer %s\n Peer : ", s.Opts.ListenAddr, msg.Key, from)
	n, err := s.Store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	fmt.Printf("[%s] stored file with key (%s) of size (%d) bytes from peer %s\n", s.Opts.ListenAddr, msg.Key, n, from)

	peer.(*p2p.TCPPeer).Wg.Done() // Signal that the stream is done

	return nil
}

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
	_, err := fs.Store.Write(key, r)
	return err
}
