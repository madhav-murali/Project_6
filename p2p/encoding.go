package p2p

import "io"

type Decoder interface {
	// Decode decodes data from the connection into a message.
	Decode(io.Reader, any) error
}
