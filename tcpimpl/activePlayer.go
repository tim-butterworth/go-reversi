package tcpimpl

import (
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

func (player *ActivePlayer) notifyOfGameStart() error {
	_, err := player.connection.Write([]byte("Game has started and you are on side " + player.side))
	return err
}

func (player *ActivePlayer) listenForPlayerInput() {
	for {
		data := make([]byte, 1024)
		for {
			offset, err := player.connection.Read(data)
			if err != nil {
				fmt.Errorf("Error reading data %s", err.Error())
			}

			if offset < 1 {
				fmt.Errorf("Error reading data %s", "connection may be closed")
				break
			}

			fmt.Println(string(data[:offset]))
			fmt.Println(player.side)

			player.commandChannel <- InfrastructureCommand{
				ResponseId:  player.ResponseId,
				CoreCommand: core.GameCommand{},
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
	fmt.Println("Starting player for side: " + player.side)
	err := player.notifyOfGameStart()
	if err != nil {
		panic(err.Error())
	}

	go player.listenForPlayerInput()
	go writeOutput(player.outputChannel, player.connection)
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
