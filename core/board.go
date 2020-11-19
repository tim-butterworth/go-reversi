package core

import (
	"github.com/google/uuid"
	"fmt"
)

type Player string
const (
	BLACK Player = "BLACK"
	WHITE Player = "WHITE"
)

func (player Player) String() string {
	switch player {
	case BLACK:
		return "BLACK"
	case WHITE:
		return "WHITE"
	default:
		return "UNKNOWN"
	}
}

type MoveCoordinate struct {
	X int
	Y int
}

type GameCommand struct {
	GameSequenceId uuid.UUID
	Coordinate MoveCoordinate
	Player Player
}

type GameEvent struct {
	GameSequenceId uuid.UUID
}

type MoveResponse struct {
	MoveApplied GameCommand
}

type Board struct {
	ActivePlayer Player
	moves []GameCommand
}

func PrintPlayers() {
	fmt.Println(BLACK)
	fmt.Println(WHITE)

	var activePlayer Player
	activePlayer = "Hi there"
	board := Board{
		ActivePlayer: activePlayer,
	}

	fmt.Println(board.ActivePlayer)
}