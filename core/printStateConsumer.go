package core

import "fmt"

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

func (consumer printStateConsumer) StateUpdated(gameState GameState) {
	board := gameState.Board
	edge := gameState.Edge
	possibleMoves := gameState.PossibleMoves

	fmt.Println()
	bounds := sequence(0, 8)
	for y := range bounds {
		rowString := ""
		for x := range bounds {
			next := "[?]"
			coordinate := Coordinate{X: x, Y: y}
			owner := board[coordinate]

			if possibleMoves.moves[coordinate] {
				next = "[ ]"
			} else if edge[coordinate] {
				next = " e "
			} else if owner == nil {
				next = "[-]"
			} else if owner.OwnedBy(BLACK) {
				next = "[B]"
			} else if owner.OwnedBy(WHITE) {
				next = "[W]"
			}

			rowString = rowString + next
		}
		fmt.Println(rowString)
	}
}

func NewPrintStateConsumer() StateUpdateConsumer {
	return printStateConsumer{}
}
