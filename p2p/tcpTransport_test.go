package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	listenAddress := "4000"
	tr := NewTCPTransport(listenAddress)
	assert.Equal(t, tr.ListenAddress, listenAddress, "Listen address should match")

	assert.Nil(t, tr.ListenAndAccept(), "Should be able to start listening without error")

}

// This function is a placeholder for testing the TCPTransport implementation.
// It should create an instance of TCPTransport, start it, and test its functionality.
// For example, you could create a listener, connect to it, and verify that connections
// are handled correctly. You might also want to test error handling and edge cases.
// You can use the testing package to create test cases and assertions.
