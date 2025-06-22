package main

import (
	"Project_6/p2p"
	"fmt"
	"log"
)

func main() {
	conf := p2p.TCPTransportOpts{
		ListenAddress: ":1000",
		ShakeHands:    p2p.NOPHandshake,     // Default handshake function, can be replaced with a custom one
		Decoder:       p2p.DefaultDecoder{}, // Replace with an actual decoder implementation
		OnPeerConnect: func(peer p2p.Peer) error {
			fmt.Printf("Doing some logic")

			return nil // No error, connection successful
		},
	}
	tr := p2p.NewTCPTransport(conf)

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("Received message: %+v", msg)
		}
	}()
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatalf("Failed to start TCP transport: %v", err)
	}

	select {} // Keep the main goroutine running to accept connections
	//fmt.Println("Listening on  osty yarn!")
}
