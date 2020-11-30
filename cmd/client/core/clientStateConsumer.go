package core

import (
	"bufio"
	"os"
	"strings"
	"fmt"
	"strconv"
	reversi_core "reversi/core"
)

type clientStateConsumer struct {
	side        reversi_core.Player
	moveChannel chan<- reversi_core.Coordinate
}

func (consumer *clientStateConsumer) StateUpdated(gameState reversi_core.GameState) {
	if gameState.PlayerTurn == consumer.side {
		fmt.Println("Your turn!")

		selectionMap := make(map[string]reversi_core.Coordinate)

		index := 1
		for move := range gameState.MoveOptions() {
			indexString := strconv.Itoa(index)
			selectionMap[indexString] = move

			fmt.Printf("possible move: %d -> (%d, %d)\n", index, move.X, move.Y)
			index = index + 1
		}

		reader := bufio.NewReader(os.Stdin)

		for {
			fmt.Print("please enter a move -> ")
			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, "\n", "", -1)

			if selected, found := selectionMap[text]; found {
				consumer.moveChannel <- selected
				break
			} else {
				fmt.Printf("%s is not a valid move selecte\n", text)
				fmt.Println("Please make another selection")
			}
		}
	} else {
		fmt.Println("Other player's turn!")
	}
}

func NewClientStateConsumer(side reversi_core.Player, moveChannel chan<- reversi_core.Coordinate) reversi_core.StateUpdateConsumer {
	return &clientStateConsumer{
		side:        side,
		moveChannel: moveChannel,
	}
}
