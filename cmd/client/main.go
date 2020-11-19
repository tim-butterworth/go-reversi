package main

import (
	"fmt"
	"net"
)

func listen(connection net.Conn, c chan<- bool, ended chan<- bool) {
	buffer := make([]byte, 1024)
	
	signaled := false
	for {
		readBytes, err := connection.Read(buffer)
		if err != nil {
			fmt.Errorf("failed to connect: %s", err.Error())
		}

		if readBytes < 1 {
			fmt.Println("Connection must have been closed :(")
			break
		}

		fmt.Println(string(buffer[:readBytes]))

		if !signaled {
			c <- true
			signaled = true
		}
	}

	ended <- true
}

func reply(connection net.Conn, c <-chan bool) {
	started := <-c

	fmt.Println("The Game is afoot!")
	if started {
		connection.Write([]byte("Thanks!"))
	}
}

func main() {
	connection, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		fmt.Errorf("failed to connect: %s", err.Error())
	}

	startedChannel := make(chan bool)
	endedChannel := make(chan bool)

	go listen(connection, startedChannel, endedChannel)
	go reply(connection, startedChannel)

	<- endedChannel
}