package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	// Decode decodes data from the connection into a message.
	Decode(io.Reader, *Message) error
}

type GOBDecoder struct{}

func (d GOBDecoder) Decode(r io.Reader, msg *Message) error {
	// Implement GOB decoding logic here
	// For example, you can use encoding/gob package to decode the message
	// return gob.NewDecoder(r).Decode(msg)
	return gob.NewDecoder(r).Decode(msg) // Placeholder, replace with actual decoding logic
}

type DefaultDecoder struct{}

func (d DefaultDecoder) Decode(r io.Reader, msg *Message) error {
	// No operation decoder, does nothing
	buf := make([]byte, 1024) // Placeholder buffer, adjust size as needed
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	//fmt.Println(string(buf[:n])) // Print the raw data read from the connection
	msg.Payload = buf[:n] // Assign the read data to the message payload
	return nil
}
