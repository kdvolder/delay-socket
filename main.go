// main.go
package main

import (
	"delay-socket/delayed_writer"
	"flag"
	"fmt"
	"io"
	"net"
	"time"
)

func main() {
	localPort := flag.Int("l", 5000, "local port to listen")
	remoteAddress := flag.String("r", "localhost:8000", "Remote address to forward to in format 'host:port'")
	flag.Parse()

	localAddress := fmt.Sprintf("localhost:%d", *localPort)

	fmt.Printf("Start port forwarder service (%d => %s)\n", *localPort, *remoteAddress)

	// Handle panics raised from the server
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("[CRITICAL] encountered a critical error, recovering from panic, error trace: %v", e)
		}
	}()

	listener, err := net.Listen("tcp", localAddress)
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	// Handler listening func
	for {
		localConnection, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		// Handle the actual forwarding to the remote
		go handlePortForward(localAddress, *remoteAddress, localConnection)
	}
}

func handlePortForward(localAddress string, remoteAddress string, local net.Conn) {
	fmt.Printf("forwarding connection %s => %s\n", localAddress, remoteAddress)

	remoteConnectionForwarded, err := net.Dial("tcp", remoteAddress)
	if err != nil {
		panic(err)
	}

	delayedRemoteWriter := delayed_writer.New(remoteConnectionForwarded, time.Millisecond*1000)

	// Ensure the local gets the response data from the remote
	go func() { io.Copy(local, remoteConnectionForwarded) }()
	// Ensure the remote gets the request data from the local
	go func() { io.Copy(delayedRemoteWriter, local) }()
}
