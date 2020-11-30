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

func listenForCommands(commandChannel <-chan InfrastructureCommand, players []*ActivePlayer) {
	factory := responderFactory{players: players}

	brain := core.NewGameBrain(factory.getSuccessInstance())
	brain.Initialize(factory.getInstance(players[0].ResponseId))

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
	CoreCommand core.Command
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
