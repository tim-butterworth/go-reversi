package main

import (
	"log"
	"net"
	"reversi/tcpimpl"
)

func listen(listener net.Listener) {
	pendingGame := tcpimpl.NewPendingGame()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("unable to accept connections: %s", err.Error())
		}

		pendingGame.AddPlayer(conn)

		if pendingGame.IsFull() {
			activeGame := tcpimpl.NewActiveGame(pendingGame)

			go activeGame.Start()
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("unable to start server: %s", err.Error())
	}

	listen(listener)
}
