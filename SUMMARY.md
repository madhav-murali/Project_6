# Project Summary (develop branch)

This document provides a summary of the project's workflow and implemented features as of the `develop` branch.

## Workflow

The application is a basic peer-to-peer (P2P) node built in Go. The workflow is as follows:

1.  **Initialization**: The application initializes a `TCPTransport` with a set of options, including a listen address, a handshake function, a message decoder, and a callback function for when a peer connects.

2.  **Listening for Connections**: The `TCPTransport` listens for incoming TCP connections on the specified address.

3.  **Accepting Connections**: When a new peer connects, the server accepts the connection and handles it. The `OnPeerConnect` callback is invoked, which currently just prints a message.

4.  **Message Consumption**: A separate goroutine continuously consumes messages from the transport's channel and prints them to the console.

## Implemented Features

*   **Peer-to-Peer Networking**: The core of the application is a P2P network based on TCP.
*   **TCP Transport**: A custom TCP transport layer handles the low-level details of peer-to-peer communication.
*   **Configurable Transport**: The TCP transport can be configured with different options, such as the listen address and handshake function.
*   **Message Decoding**: A `Decoder` is used to decode incoming messages from peers.
*   **Asynchronous Message Processing**: Messages are processed asynchronously using Go channels.
