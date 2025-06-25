package main

import (
	"Project_6/p2p"
	"log"
)

// makeServer initializes a new FileServer with the given listen address and root directory.
// It sets up the TCP transport and the path transformation function.
func makeServer(ListenAddr string, nodes ...string) *FileServer {

	tcpTransport := p2p.NewTCPTransport(p2p.TCPTransportOpts{
		ListenAddress: ListenAddr,
		ShakeHands:    p2p.NOPHandshake,     // Default handshake function
		Decoder:       p2p.DefaultDecoder{}, // Replace with an actual decoder implementation
	})

	fsOpts := FileServerOpts{
		ListenAddr:        ListenAddr,
		StorageRoot:       ListenAddr + "_network", // Use the listen address as the root directory
		PathTransformFunc: CASPathTransformFunc,    // Use the custom path transform function
		Transport:         tcpTransport,
		BootStrapNodes:    nodes, // Example bootstrap nodes
	}

	fs := NewFileServer(fsOpts)

	tcpTransport.OnPeerConnect = fs.AddPeer

	return NewFileServer(fsOpts)
}

// func main() {
// 	fs := makeServer()

// 	go func() {
// 		time.Sleep(time.Second * 3)
// 		fs.Stop() // Stop the file server after 3 seconds
// 		log.Println("File server stopped")
// 	}()

// 	if err := fs.Start(); err != nil {
// 		log.Fatalf("Failed to start file server: %v", err)
// 	}

// 	//select {} // Keep the main goroutine running to accept connections

// }

func main() {
	fs1 := makeServer(":3000", "")
	fs2 := makeServer(":4000", ":3000")
	go func() {
		log.Fatal(fs1.Start())
	}()
	fs2.Start() // Start the second server
	// Start another server on a different port
	// This can be used fofr testing or as a bootstrap node

}
