package core

import (
	"fmt"
	reversi_core "reversi/core"
)

type printStateConsumer struct{}

func sequence(min int, max int) []int {
	if max <= min {
		return make([]int, 0)
	}

	result := make([]int, max-min)
	current := min
	index := 0
	for {
		if !(current < max) {
			break
		}

		result[index] = current

		index++
		current = min + index
	}

	return result
}

func (consumer printStateConsumer) StateUpdated(gameState reversi_core.GameState) {
	board := gameState.Board
	edge := gameState.Edge
	possibleMoves := gameState.MoveOptions()

	fmt.Println()
	bounds := sequence(0, 8)

	header := "   "
	for x := range bounds {
		header = header + fmt.Sprintf(" %d ", x)
	}
	fmt.Println(header)

	for y := range bounds {
		rowString := fmt.Sprintf(" %d ", y)
		for x := range bounds {
			next := "[?]"
			coordinate := reversi_core.Coordinate{X: x, Y: y}
			owner := board[coordinate]

			if possibleMoves[coordinate] {
				next = " - "
			} else if edge[coordinate] {
				next = " - "
			} else if owner == nil {
				next = " - "
			} else if owner.OwnedBy(reversi_core.BLACK) {
				next = " X "
			} else if owner.OwnedBy(reversi_core.WHITE) {
				next = " 0 "
			}

			rowString = rowString + next
		}
		fmt.Println(rowString)
		fmt.Println()
	}
}

func NewPrintStateConsumer() reversi_core.StateUpdateConsumer {
	return printStateConsumer{}
}
