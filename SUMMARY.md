# Project Summary

This document provides a summary of the project's workflow and implemented features.

## Workflow

The application is a peer-to-peer (P2P) networking tool built in Go. The workflow is as follows:

1.  **Initialization**: The application starts by initializing a `TCPTransport` with a set of options, including a listen address, a handshake function, a message decoder, and a callback function for when a peer connects.

2.  **Listening for Connections**: The `TCPTransport` listens for incoming TCP connections on the specified address.

3.  **Accepting Connections**: When a new peer connects, the server accepts the connection and handles it in a new goroutine.

4.  **Handshake**: A handshake is performed with the new peer to ensure it's a valid participant in the network.

5.  **Peer Connection Callback**: If the handshake is successful, the `OnPeerConnect` callback is invoked, allowing for custom logic to be executed.

6.  **Message Handling**: The application enters a loop where it decodes messages from the peer. Decoded messages are sent to a channel for asynchronous processing.

7.  **Message Consumption**: A separate goroutine consumes messages from the channel and processes them.

## Implemented Features

*   **Peer-to-Peer Networking**: The core of the application is a P2P network based on TCP.
*   **TCP Transport**: A custom TCP transport layer handles the low-level details of peer-to-peer communication.
*   **Configurable Transport**: The TCP transport can be configured with different options, such as the listen address and handshake function.
*   **Handshake Mechanism**: A simple handshake mechanism is in place to verify peers.
*   **Message Decoding**: A `Decoder` is used to decode incoming messages from peers.
*   **Asynchronous Message Processing**: Messages are processed asynchronously using Go channels.
*   **Peer Representation**: The `TCPPeer` struct represents a connected peer in the network.
