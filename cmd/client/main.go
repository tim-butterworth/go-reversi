package main

import (
	"encoding/json"
	"fmt"
	"errors"
	"net"
	"reversi/cmd/client/core"
	reversi_core "reversi/core"
	"reversi/tcpimpl"
)

func convertToInt(raw interface{}) (error, int) {
	t := fmt.Sprintf("%T", raw)

	if t == "float64" {
		return nil, int(raw.(float64))
	}

	return errors.New("failed to convert the value to an int :("), 0
}

func convertToMovedEvent(rawEvent reversi_core.Event) (error, reversi_core.Event) {
	rawMap := rawEvent.Data.(map[string]interface{})

	xErr, x := convertToInt(rawMap["X"])
	yErr, y := convertToInt(rawMap["Y"])

	emptyEvent := reversi_core.Event{}
	if xErr != nil {
		return xErr, emptyEvent
	}
	if yErr != nil {
		return yErr, emptyEvent
	}

	return nil, reversi_core.NewMoveEvent(reversi_core.Coordinate{X: x, Y: y})
}

func listen(connection net.Conn, c chan<- bool, ended chan<- bool, moveChannel chan<- reversi_core.Coordinate) {
	buffer := make([]byte, 1024)

	signaled := false
	gameStarted := false

	gameState := reversi_core.NewGameEventAggregator()
	gameState.Register(core.NewPrintStateConsumer())

	for {
		readBytes, err := connection.Read(buffer)
		if err != nil {
			fmt.Errorf("failed to connect: %s", err.Error())
		}

		if readBytes < 1 {
			fmt.Println("Connection must have been closed :(")
			break
		}

		if !gameStarted {
			sideAssigned := tcpimpl.SideAssigned{}
			json.Unmarshal(buffer[:readBytes], &sideAssigned)

			gameStarted = true
			fmt.Printf("Game has started and you have been assigned side ->  [%s]\n", sideAssigned.Side)

			gameState.Register(core.NewClientStateConsumer(sideAssigned.Side, moveChannel))
		} else {
			event := reversi_core.Event{}

			json.Unmarshal(buffer[:readBytes], &event)

			if event.EventType == reversi_core.MOVED {
				err, event = convertToMovedEvent(event)
				if err != nil {
					fmt.Errorf("Failed to convert to MovedEvent %s\n", err.Error())
					continue
				}
			}

			gameState.SendEvent(event)
		}

		if !signaled {
			c <- true
			signaled = true
		}
	}

	ended <- true
}

func reply(connection net.Conn, c <-chan bool, moveChannel <-chan reversi_core.Coordinate) {
	_ = <-c

	for {
		move := <-moveChannel
		data, err := json.Marshal(move)
		if err != nil {
			fmt.Errorf("Error serializing move %s\n", err.Error())
		}

		_, err = connection.Write(data)
		if err != nil {
			fmt.Errorf("Error writing data %s", err.Error())
		}
	}
}

func main() {
	connection, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		fmt.Errorf("failed to connect: %s", err.Error())
	}

	startedChannel := make(chan bool)
	endedChannel := make(chan bool)
	moveChannel := make(chan reversi_core.Coordinate)

	go listen(connection, startedChannel, endedChannel, moveChannel)
	go reply(connection, startedChannel, moveChannel)

	_ = <-endedChannel
}
