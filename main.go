package main

import (
	"Project_6/p2p"
	"fmt"
	"log"
)

func main() {
	conf := p2p.TCPTransportOpts{
		ListenAddress: ":1000",
		ShakeHands:    p2p.NOPHandshake, // Default handshake function, can be replaced with a custom one
		Decoder:       nil,              // Replace with an actual decoder implementation
	}
	tr := p2p.NewTCPTransport(conf)

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatalf("Failed to start TCP transport: %v", err)
	}

	select {} // Keep the main goroutine running to accept connections
	fmt.Println("Listening on  osty yarn!")
}
