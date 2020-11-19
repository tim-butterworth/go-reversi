package core

import (
	"fmt"
)

type Responder interface {
	Respond(message string)
	RespondAll(message string)
	NotifyActivePlayer(message string, activePlayerSide Player)
	NotifyInactivePlayer(message string, activePlayerSide Player)
}

type CellClaim interface {
	OwnedBy(player Player) bool
}
type ownedByBlack struct{}

func (owner ownedByBlack) OwnedBy(player Player) bool {
	return player == BLACK
}

type ownedByWhite struct{}

func (owner ownedByWhite) OwnedBy(player Player) bool {
	return player == WHITE
}

type notOwned struct{}

func (owner notOwned) OwnedBy(player Player) bool {
	return false
}

type GameBrain struct {
	GameState    [][]CellClaim
	ActivePlayer Player
}

func (brain *GameBrain) ExecuteCommand(command GameCommand, responder Responder) {
	responder.RespondAll("Hi from the brain of the game!")
	responder.Respond("Special secret message")

	responder.NotifyActivePlayer("Your turn!", brain.ActivePlayer)
	responder.NotifyInactivePlayer("Not your turn just yet!", brain.ActivePlayer)
}

func stringValue(b bool) string {
	if b {
		return "t"
	}
	return "f"
}

func getNeighborhood(c Coordinate) []Coordinate {
	result := make([]Coordinate, 9)

	x := c.X
	y := c.Y

	fmt.Printf("x -> %d, y -> %d\n", x, y)
	values := []int{-1, 0, 1}

	itterator := NewCounter([][]int{values, values})
	for i := range result {
		v := itterator.Next()
		result[i] = Coordinate{X: x + v[0], Y: y + v[1]}
	}

	return result
}

func getInitialBoard() [][]CellClaim {
	fullRange := make([]CellClaim, 64)
	result := make([][]CellClaim, 8)

	noOwner := notOwned{}
	ownedByBlack := ownedByBlack{}
	ownedByWhite := ownedByWhite{}

	for i := range fullRange {
		fullRange[i] = noOwner
	}

	for row := range result {
		min := row * 8
		max := min + 8

		result[row] = fullRange[min:max]
		row++
	}

	result[3][3] = ownedByBlack
	result[4][4] = ownedByBlack

	result[4][3] = ownedByWhite
	result[3][4] = ownedByWhite

	used := make(map[Coordinate]bool)
	edge := make(map[Coordinate]bool)

	used[Coordinate{X: 3, Y: 3}] = true
	used[Coordinate{X: 4, Y: 4}] = true
	used[Coordinate{X: 4, Y: 3}] = true
	used[Coordinate{X: 3, Y: 4}] = true

	for k, _ := range used {
		neighborhood := getNeighborhood(k)
		for _, v := range neighborhood {
			if !used[v] {
				edge[v] = true
			}
		}
	}

	for x, row := range result {
		rowString := ""
		for y, owner := range row {
			next := "_"
			if owner.OwnedBy(BLACK) {
				next = "B"
			} else if owner.OwnedBy(WHITE) {
				next = "W"
			} else if edge[Coordinate{X: x, Y: y}]{
				next = "E"
			}

			rowString = rowString + next
		}
		fmt.Println(rowString)
	}

	return result
}

func NewGameBrain() GameBrain {
	return GameBrain{
		GameState:    getInitialBoard(),
		ActivePlayer: BLACK,
	}
}
