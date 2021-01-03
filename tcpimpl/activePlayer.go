package tcpimpl

import (
	"encoding/json"
	"fmt"
	"reversi/core"

	"github.com/google/uuid"
)

type ActivePlayer struct {
	side           core.Player
	connection     Connection
	ResponseId     uuid.UUID
	commandChannel chan<- InfrastructureCommand
	outputChannel  chan string
}

func (player *ActivePlayer) RespondsTo(side core.Player) bool {
	return player.side == side
}

type SideAssigned struct {
	Side core.Player
	Id   uuid.UUID
}

type SideAssignedEvent struct {
	EventType string
	Data      SideAssigned
}

func (player *ActivePlayer) notifyOfGameStart() error {
	sideAssigned := SideAssigned{
		Side: player.side,
		Id:   player.ResponseId,
	}
	sideAssignedEvent := SideAssignedEvent{
		EventType: "SIDE_ASSIGNED",
		Data:      sideAssigned,
	}
	data, err := json.Marshal(sideAssignedEvent)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	player.Notify(string(data))
	return nil
}

func (player *ActivePlayer) listenForPlayerInput() {
	defer player.connection.Close()
	for {
		coordinate := core.Coordinate{}
		err := player.connection.ReadJson(&coordinate)
		if err != nil {
			fmt.Printf("Error reading json -> %s\n", err.Error())
			break
		}

		fmt.Printf("Received Coordinate from client %s with value (%d, %d)\n", player.side, coordinate.X, coordinate.Y)

		player.commandChannel <- InfrastructureCommand{
			ResponseId:  player.ResponseId,
			CoreCommand: core.NewMoveCommand(player.side, coordinate),
		}
	}
}

func writeOutput(outputChannel <-chan string, connection Connection) {
	for {
		message := <-outputChannel

		connection.Message(message)
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
		fmt.Printf("Giving up -> %s\n", err.Error())
		panic(err.Error())
	}
}

func NewActivePlayer(side core.Player, connection Connection, commandChannel chan<- InfrastructureCommand) *ActivePlayer {
	return &ActivePlayer{
		side:           side,
		connection:     connection,
		commandChannel: commandChannel,
		ResponseId:     uuid.New(),
		outputChannel:  make(chan string),
	}
}
