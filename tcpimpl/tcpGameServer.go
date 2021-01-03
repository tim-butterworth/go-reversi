package tcpimpl

import (
	"fmt"

	"github.com/google/uuid"
)

type PendingGame struct {
	players []PlayerConnection
}

func (pending *PendingGame) addPlayer(connection Connection) error {
	if len(pending.players) < 2 {
		uuid, err := uuid.NewUUID()
		if err != nil {
			return err
		}

		playerConnection := PlayerConnection{
			Connection: connection,
			Id:         uuid,
		}
		pending.players = append(pending.players, playerConnection)
	}

	return nil
}

func (pending *PendingGame) IsFull() bool {
	return len(pending.players) == 2
}

func (pending *PendingGame) stillAcceptingPlayers() bool {
	return len(pending.players) < 2
}

type Connection interface {
	Error(message string)
	Message(message string)
	ReadJson(container interface{}) error
	Close()
}

func (pendingGame *PendingGame) AddPlayer(playerConnection Connection) {
	if pendingGame.stillAcceptingPlayers() {
		addPlayerErr := pendingGame.addPlayer(playerConnection)
		if addPlayerErr != nil {
			defer playerConnection.Close()
			playerConnection.Error("failed to generate an id")
			fmt.Errorf("failed to add the player to a pending game: %s", addPlayerErr.Error())
		}
	} else {
		defer playerConnection.Close()
		playerConnection.Message("Maximum players reached")
	}
}

func NewPendingGame() PendingGame {
	var players []PlayerConnection
	return PendingGame{
		players: players,
	}
}
