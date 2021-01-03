package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reversi/tcpimpl"
)

func tcpMessage(conn net.Conn, message string) {
	_, err := conn.Write([]byte(message))
	if err != nil {
		log.Fatalf("failed to send message: %s", err.Error())
	}
}

type TCPPendingPlayer struct {
	connection net.Conn
}

func (player *TCPPendingPlayer) Error(message string) {
	tcpMessage(player.connection, message)
}

func (player *TCPPendingPlayer) Message(message string) {
	tcpMessage(player.connection, message)
}

func (player *TCPPendingPlayer) ReadJson(container interface{}) error {
	buffer := make([]byte, 1024)
	i, err := player.connection.Read(buffer)
	if err != nil {
		fmt.Errorf("Error reading bytes: %s\n", err.Error())
		return err
	}

	if i == len(buffer)-1 {
		fmt.Errorf("The message may be too large, need to reconsider the implementation, maybe need to do fancy looping if the message is bigger than 1024")
	}

	return json.Unmarshal(buffer[:i], &container)
}

func (player *TCPPendingPlayer) Close() {
	player.connection.Close()
}

func listen(listener net.Listener) {
	pendingGame := tcpimpl.NewPendingGame()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("unable to accept connections: %s", err.Error())
		}

		pendingGame.AddPlayer(&TCPPendingPlayer{connection: conn})

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
