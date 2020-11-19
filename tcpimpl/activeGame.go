package tcpimpl

import (
	"fmt"
	"reversi/core"

	"github.com/google/uuid"
)

type ActiveGame interface {
	Start() error
}

type activeGameImpl struct {
	players     []*ActivePlayer
	moveChannel <-chan InfrastructureCommand
}

func asString(b bool) string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

type infraResponder struct {
	players    []*ActivePlayer
	responseId uuid.UUID
}

func (responder infraResponder) RespondAll(message string) {
	for _, player := range responder.players {
		player.Notify(message)
	}
}

func (responder infraResponder) Respond(message string) {
	for _, player := range responder.players {
		if player.ResponseId == responder.responseId {
			player.Notify(message + " -> " + responder.responseId.String())
		}
	}
}

func (responder infraResponder) NotifyActivePlayer(message string, activePlayerSide core.Player) {
	for _, player := range responder.players {
		if player.side == activePlayerSide {
			player.Notify(message + " -> " + "for player of side black")
		}
	}
}

func (responder infraResponder) NotifyInactivePlayer(message string, activePlayerSide core.Player) {
	for _, player := range responder.players {
		if player.side != activePlayerSide {
			player.Notify(message + " -> " + "for player of side white")
		}
	}
}

type responderFactory struct {
	players []*ActivePlayer
}

func (factory responderFactory) getInstance(responseId uuid.UUID) infraResponder {
	return infraResponder{
		players:    factory.players,
		responseId: responseId,
	}
}

func listenForCommands(commandChannel <-chan InfrastructureCommand, players []*ActivePlayer) {
	factory := responderFactory{players: players}

	brain := core.NewGameBrain()

	for {
		command := <-commandChannel

		brain.ExecuteCommand(
			command.CoreCommand,
			factory.getInstance(command.ResponseId),
		)
	}
}

func (activeGame *activeGameImpl) Start() error {
	fmt.Println("Starting the game!")

	for _, player := range activeGame.players {
		player.start()
	}

	go listenForCommands(activeGame.moveChannel, activeGame.players)

	return nil
}

type InfrastructureCommand struct {
	CoreCommand core.GameCommand
	ResponseId  uuid.UUID
}

func NewActiveGame(pendingGame PendingGame) ActiveGame {
	players := pendingGame.players
	blackPlayer := players[0]
	whitePlayer := players[1]

	gameCommandChannel := make(chan InfrastructureCommand)

	blackGamePlayer := NewActivePlayer(core.BLACK, blackPlayer.Connection, gameCommandChannel)
	whiteGamePlayer := NewActivePlayer(core.WHITE, whitePlayer.Connection, gameCommandChannel)

	gamePlayers := make([]*ActivePlayer, 2)
	gamePlayers[0] = blackGamePlayer
	gamePlayers[1] = whiteGamePlayer

	return &activeGameImpl{
		players:     gamePlayers,
		moveChannel: gameCommandChannel,
	}
}
