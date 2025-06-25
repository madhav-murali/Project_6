package p2p

type RPC struct {
	// From    net.Addr // The address of the peer sending the message
	// To      net.Addr // The address of the peer receiving the message
	Key     string      // The key for the data being sent
	Payload interface{} // The actual data being sent in the message
}
