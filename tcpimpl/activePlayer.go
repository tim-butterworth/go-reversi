package tcpimpl

import (
	"encoding/json"
	"fmt"
	"net"
	"reversi/core"

	"github.com/google/uuid"
)

type ActivePlayer struct {
	side           core.Player
	connection     net.Conn
	ResponseId     uuid.UUID
	commandChannel chan<- InfrastructureCommand
	outputChannel  chan string
}

func (player *ActivePlayer) RespondsTo(side core.Player) bool {
	return player.side == side
}

type SideAssigned struct {
	Side core.Player
}

func (player *ActivePlayer) notifyOfGameStart() error {
	sideAssigned := SideAssigned{Side: player.side}
	data, err := json.Marshal(sideAssigned)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	player.Notify(string(data))
	return nil
}

func (player *ActivePlayer) listenForPlayerInput() {
	for {
		data := make([]byte, 1024)
		for {
			offset, err := player.connection.Read(data)
			if err != nil {
				fmt.Errorf("Error reading data %s\n", err.Error())
			}

			if offset < 1 {
				fmt.Errorf("Error reading data %s", "connection may be closed\n")
				break
			}

			coordinate := core.Coordinate{}
			json.Unmarshal(data[:offset], &coordinate)

			fmt.Printf("Received Coordinate from client %s with value (%d, %d)\n", player.side, coordinate.X, coordinate.Y)

			player.commandChannel <- InfrastructureCommand{
				ResponseId:  player.ResponseId,
				CoreCommand: core.NewMoveCommand(player.side, coordinate),
			}
		}
	}
}

func writeOutput(outputChannel <-chan string, connection net.Conn) {
	for {
		message := <-outputChannel

		_, err := connection.Write([]byte(message + "\n"))
		if err != nil {
			panic(err.Error())
		}
	}
}

func (player *ActivePlayer) Notify(message string) {
	player.outputChannel <- message
}

func (player *ActivePlayer) RespondsFor(responseId uuid.UUID) bool {
	return player.ResponseId == responseId
}

func (player ActivePlayer) start() {
	go player.listenForPlayerInput()
	go writeOutput(player.outputChannel, player.connection)

	fmt.Println("Starting player for side: " + player.side)
	err := player.notifyOfGameStart()
	if err != nil {
		panic(err.Error())
	}
}

func NewActivePlayer(side core.Player, connection net.Conn, commandChannel chan<- InfrastructureCommand) *ActivePlayer {
	return &ActivePlayer{
		side:           side,
		connection:     connection,
		commandChannel: commandChannel,
		ResponseId:     uuid.New(),
		outputChannel:  make(chan string),
	}
}
