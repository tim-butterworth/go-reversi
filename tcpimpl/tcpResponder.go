package tcpimpl

import (
	"encoding/json"
	"fmt"
	"reversi/core"

	"github.com/google/uuid"
)

type Coordinate struct {
	X int
	Y int
}

type tcpResponder struct {
	infraResponder infraResponder
}

func (responder *tcpResponder) moveFailure() {
	fmt.Println("Move failed")

	responder.infraResponder.respond("Move Failed")
}

func newTcpResponder(infraResponder infraResponder) *tcpResponder {
	return &tcpResponder{infraResponder: infraResponder}
}

func (responder *tcpResponder) InvalidCommand(command core.Command) {
	responder.infraResponder.respond("Command Rejected... so")
}

type infraResponder struct {
	players    []*ActivePlayer
	responseId uuid.UUID
}

func (responder *infraResponder) respond(message string) {
	for _, player := range responder.players {
		if player.ResponseId == responder.responseId {
			player.Notify(message + " -> " + responder.responseId.String())
		}
	}
}

func (responder *infraResponder) notifyActivePlayer(message string, activePlayerSide core.Player) {
	for _, player := range responder.players {
		if player.side == activePlayerSide {
			player.Notify(message + " -> " + "for player of side black")
		}
	}
}

func (responder *infraResponder) notifyInactivePlayer(message string, activePlayerSide core.Player) {
	for _, player := range responder.players {
		if player.side != activePlayerSide {
			player.Notify(message + " -> " + "for player of side white")
		}
	}
}

type responderFactory struct {
	players []*ActivePlayer
}

func (factory responderFactory) getInstance(responseId uuid.UUID) core.CommandRejectHandler {
	infraResponder := infraResponder{
		players:    factory.players,
		responseId: responseId,
	}

	return newTcpResponder(infraResponder)
}

type SuccessResponder struct {
	players []*ActivePlayer
}

func (responder SuccessResponder) SendEvent(event core.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		fmt.Println("Got an erorr :(")
	}

	fmt.Println(string(data))

	for _, player := range responder.players {
		player.Notify(string(data))
	}
}

func (factory responderFactory) getSuccessInstance() core.EventConsumer {
	successResponder := SuccessResponder{
		players: factory.players,
	}

	return &successResponder
}
