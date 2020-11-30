package core_test

import (
	"fmt"
	"reversi/core"
	"testing"
)

type TestStateUpdateConsumer struct {
	state     core.GameState
	wasCalled bool
}

func (consumer *TestStateUpdateConsumer) StateUpdated(gameState core.GameState) {
	consumer.wasCalled = true
	consumer.state = gameState
}

func newTestStateUpdateConsumer() TestStateUpdateConsumer {
	var state core.GameState
	return TestStateUpdateConsumer{
		state:     state,
		wasCalled: false,
	}
}

func printBoard(board map[core.Coordinate]core.CellClaim) {
	for cell, owner := range board {
		var side string
		if owner.OwnedBy(core.WHITE) {
			side = fmt.Sprintf("%s", core.WHITE)
		} else if owner.OwnedBy(core.BLACK) {
			side = fmt.Sprintf("%s", core.BLACK)
		} else {
			side = "?"
		}

		fmt.Printf("(%d, %d) -> %s\n", cell.X, cell.Y, side)
	}
}

func shouldBe(t *testing.T, cellClaim core.CellClaim, side core.Player, failureMessage string) {
	if !cellClaim.OwnedBy(side) {
		t.Errorf("[Failed] %s", failureMessage)
	}
}
func shouldBeBlack(t *testing.T, coordinate core.Coordinate, board map[core.Coordinate]core.CellClaim) {
	shouldBe(t, board[coordinate], core.BLACK, fmt.Sprintf("(%d, %d) should have been %s", coordinate.X, coordinate.Y, core.BLACK))
}
func shouldBeWhite(t *testing.T, coordinate core.Coordinate, board map[core.Coordinate]core.CellClaim) {
	shouldBe(t, board[coordinate], core.WHITE, fmt.Sprintf("(%d, %d) should have been %s", coordinate.X, coordinate.Y, core.WHITE))
}

func Test_InitializeEvent_triggersGameInitialization(t *testing.T) {
	testStateUpdateConsumer := newTestStateUpdateConsumer()

	aggregator := core.NewGameEventAggregator()
	aggregator.Register(&testStateUpdateConsumer)

	aggregator.SendEvent(core.NewInitializedEvent())

	if testStateUpdateConsumer.wasCalled {
		gameState := testStateUpdateConsumer.state

		shouldBeWhite(t, core.Coordinate{X: 3, Y: 4}, gameState.Board)
		shouldBeWhite(t, core.Coordinate{X: 4, Y: 3}, gameState.Board)

		shouldBeBlack(t, core.Coordinate{X: 3, Y: 3}, gameState.Board)
		shouldBeBlack(t, core.Coordinate{X: 4, Y: 4}, gameState.Board)

		if gameState.PlayerTurn == core.WHITE {
			t.Error("Should be BLACK's turn")
		}
	} else {
		t.Error("the stateUpdateConsumer should have been called")
	}
}

func Test_MoveEvent_triggersGameStateUpdate(t *testing.T) {
	testStateUpdateConsumer := newTestStateUpdateConsumer()

	aggregator := core.NewGameEventAggregator()
	aggregator.Register(&testStateUpdateConsumer)

	aggregator.SendEvent(core.NewInitializedEvent())
	aggregator.SendEvent(core.NewMoveEvent(core.Coordinate{X: 2, Y: 4}))

	if testStateUpdateConsumer.wasCalled {
		gameState := testStateUpdateConsumer.state

		shouldBeBlack(t, core.Coordinate{X: 2, Y: 4}, gameState.Board)
		shouldBeBlack(t, core.Coordinate{X: 3, Y: 4}, gameState.Board)

		shouldBeWhite(t, core.Coordinate{X: 4, Y: 3}, gameState.Board)

		if gameState.PlayerTurn == core.BLACK {
			t.Error("Should be WHITE's turn")
		}
	} else {
		t.Error("the stateUpdateConsumer should have been called")
	}
}

func Test_MoveEvent_canPlayAWholeGame(t *testing.T) {
	testStateUpdateConsumer := newTestStateUpdateConsumer()

	aggregator := core.NewGameEventAggregator()
	aggregator.Register(&testStateUpdateConsumer)
	aggregator.Register(core.NewPrintStateConsumer())

	aggregator.SendEvent(core.NewInitializedEvent())

	moves := []core.Coordinate{
		core.Coordinate{X: 2, Y: 4},
		core.Coordinate{X: 2, Y: 5},
		core.Coordinate{X: 2, Y: 6},
		core.Coordinate{X: 1, Y: 4},
		core.Coordinate{X: 0, Y: 4},
		core.Coordinate{X: 4, Y: 5},
		core.Coordinate{X: 5, Y: 2},
		core.Coordinate{X: 4, Y: 2},
		core.Coordinate{X: 3, Y: 2},
		core.Coordinate{X: 2, Y: 1},
		core.Coordinate{X: 3, Y: 1},
		core.Coordinate{X: 4, Y: 1},
		core.Coordinate{X: 3, Y: 0},
		core.Coordinate{X: 1, Y: 5},
		core.Coordinate{X: 4, Y: 6},
		core.Coordinate{X: 2, Y: 7},
		core.Coordinate{X: 3, Y: 6},
		core.Coordinate{X: 1, Y: 3},
		core.Coordinate{X: 0, Y: 5},
		core.Coordinate{X: 3, Y: 5},
		core.Coordinate{X: 0, Y: 2},
		core.Coordinate{X: 1, Y: 2},
		core.Coordinate{X: 2, Y: 3},
		core.Coordinate{X: 2, Y: 2},
		core.Coordinate{X: 3, Y: 7},
		core.Coordinate{X: 0, Y: 3},
		core.Coordinate{X: 2, Y: 0},
		core.Coordinate{X: 6, Y: 2},
		core.Coordinate{X: 7, Y: 2},
		core.Coordinate{X: 5, Y: 3},
		core.Coordinate{X: 6, Y: 4},
		core.Coordinate{X: 5, Y: 4},
		core.Coordinate{X: 4, Y: 7},
		core.Coordinate{X: 7, Y: 4},
		core.Coordinate{X: 7, Y: 5},
		core.Coordinate{X: 5, Y: 7},
		core.Coordinate{X: 7, Y: 3},
		core.Coordinate{X: 6, Y: 3},
		core.Coordinate{X: 5, Y: 6},
		core.Coordinate{X: 5, Y: 5},
		core.Coordinate{X: 6, Y: 5},
		core.Coordinate{X: 5, Y: 0},
		core.Coordinate{X: 5, Y: 1},
		core.Coordinate{X: 4, Y: 0},
		core.Coordinate{X: 6, Y: 0},
		core.Coordinate{X: 1, Y: 7},
		core.Coordinate{X: 1, Y: 1},
		core.Coordinate{X: 1, Y: 0},
		core.Coordinate{X: 0, Y: 0},
		core.Coordinate{X: 0, Y: 1},
		core.Coordinate{X: 1, Y: 6},
		core.Coordinate{X: 0, Y: 7},
		core.Coordinate{X: 0, Y: 6},
		core.Coordinate{X: 7, Y: 6},
		core.Coordinate{X: 7, Y: 7},
		core.Coordinate{X: 6, Y: 6},
		core.Coordinate{X: 6, Y: 7},
		core.Coordinate{X: 7, Y: 1},
		core.Coordinate{X: 6, Y: 1},
		core.Coordinate{X: 7, Y: 0},
	}

	for _, move := range moves {
		aggregator.SendEvent(core.NewMoveEvent(move))
	}

	if testStateUpdateConsumer.wasCalled {
		gameState := testStateUpdateConsumer.state
		moveCount := len(gameState.Board)
		if !(moveCount == 64) {
			t.Errorf("All 64 cells should have been used, instead, %d were used", moveCount)
		}

		if gameState.PlayerTurn == core.BLACK {
			t.Error("Should be WHITE's turn")
		}
	} else {
		t.Error("the stateUpdateConsumer should have been called")
	}
}

//Figure out how to test if a player has no moves and that turn gets skipped
