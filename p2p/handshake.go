package p2p

// type HandShaker interface {
// 	// Handshake initiates a handshake with the peer.
// 	Handshake(peer Peer) error
// }
//this error is used when the handshake fails
// it can be used to indicate that the handshake was not successful, for example, if the
//var ErrInvalidHandshake = errors.New("invalid handshake")

type HandShakeFunc func(Peer) error

func NOPHandshake(Peer) error {
	// No operation handshake function, does nothing
	return nil
}
