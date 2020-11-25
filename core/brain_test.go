package core_test

import (
	"fmt"
	"reversi/core"
	"testing"
)

type boolAssertions struct {
	actual bool
	t      *testing.T
}

func boolAssertion(actual bool, expected bool, t *testing.T, message string) {
	if actual != expected {
		t.Errorf("Expected %t but got %t ::> %s\n", expected, actual, message)
	}
}

func (assertions boolAssertions) toBeTrue(message string) {
	expected := true
	actual := assertions.actual

	boolAssertion(actual, expected, assertions.t, message)
}

func (assertions boolAssertions) toBeFalse(message string) {
	expected := false
	actual := assertions.actual

	boolAssertion(actual, expected, assertions.t, message)
}

func expect(t *testing.T, b bool) boolAssertions {
	return boolAssertions{
		actual: b,
		t:      t,
	}
}

type TestResultHandler struct {
	result core.NextPlayInfo
}

func (handler *TestResultHandler) MoveSuccess(result core.MoveSuccessResult) {
	fmt.Println("Success called!")
	fmt.Println(result.NextPlayerInfo().NextPlayer())
	fmt.Println(len(result.NextPlayerInfo().Moves()))

	handler.result = result.NextPlayerInfo()
}
func (handler *TestResultHandler) MoveFailure() {
	fmt.Println("Failure called!")
}
func (handler *TestResultHandler) GameInitialized(nextPlay core.NextPlayInfo) {
	handler.result = nextPlay

	fmt.Println("Initialize called")
}

func TestIsPossibleMove_returnsTrue_forPossibleMoves(t *testing.T) {
	validCoordinates := []core.Coordinate{
		core.Coordinate{X: 2, Y: 3},
		core.Coordinate{X: 3, Y: 2},
		core.Coordinate{X: 4, Y: 5},
		core.Coordinate{X: 5, Y: 4},
	}
	for _, coordinate := range validCoordinates {
		actual := core.IsPossibleMove(core.WHITE, coordinate, core.GetInitialBoard())

		expect(t, actual).toBeTrue(fmt.Sprintf("for: (%d, %d)", coordinate.X, coordinate.Y))
	}
}
func TestIsPossibleMove_returnsFalse_forNonPossibleMoves(t *testing.T) {
	validCoordinates := []core.Coordinate{
		core.Coordinate{X: 2, Y: 2},
		core.Coordinate{X: 2, Y: 4},
		core.Coordinate{X: 4, Y: 2},
		core.Coordinate{X: 5, Y: 5},
		core.Coordinate{X: 0, Y: 0},
		core.Coordinate{X: 7, Y: 7},
		core.Coordinate{X: 8, Y: 8},
		core.Coordinate{X: 500, Y: 140},
	}
	for _, coordinate := range validCoordinates {
		actual := core.IsPossibleMove(core.WHITE, coordinate, core.GetInitialBoard())

		expect(t, actual).toBeFalse(fmt.Sprintf("for: (%d, %d)", coordinate.X, coordinate.Y))
	}
}

func TestAttemptSomeMoves(t *testing.T) {
	handler := TestResultHandler{}
	brain := core.NewGameBrain(&handler)

	brain.Initialize(&handler)

	moveList := handler.result.Moves()
	move := core.Move{Coordinate: moveList[0], Side: handler.result.NextPlayer()}
	fmt.Printf("(%d, %d)\n", move.Coordinate.X, move.Coordinate.Y)

	appliedMoves := make([]core.Move, 64)
	moveCount := 0
	for {
		brain.AttemptMove(move, &handler)
		brain.PrintGameState()

		moveResult := handler.result
		moves := moveResult.Moves()
		side := moveResult.NextPlayer()

		if len(moves) == 0 {
			break
		}

		move = core.Move{Side: side, Coordinate: moves[0]}

		appliedMoves[moveCount] = move
		moveCount++
	}

	appliedMoves = appliedMoves[0:moveCount]

	for _, move := range appliedMoves {
		fmt.Printf("Side -> %s (%d, %d)\n", move.Side, move.Coordinate.X, move.Coordinate.Y)
	}
}
