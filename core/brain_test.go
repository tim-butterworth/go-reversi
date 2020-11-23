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

type TestResultHandler struct{}

func (handler TestResultHandler) MoveSuccess(result core.MoveResult) {
	fmt.Println("Success called!")
	fmt.Println(result.Side())
	for _, move := range result.Moves() {
		fmt.Printf("(%d, %d)\n", move.X, move.Y)
	}
}
func (handler TestResultHandler) MoveFailure() {
	fmt.Println("Failure called!")
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
// BWeee---
// BW_We---
// eBBBee--
// eWBBWe--
// eeeBBe--
// --_B__--
// --e_e---
// --------

// BWeee---
// BWWWe---
// eBBWee--
// eWBBWe--
// eeeBBe--
// --_B__--
// --e_e---
// --------
func TestAttemptSomeMoves(t *testing.T) {
	brain := core.NewGameBrain()

	handler := TestResultHandler{}
	moves := []core.Move{
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 3, Y: 5}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 2, Y: 3}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 1, Y: 2}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 1, Y: 3}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 2, Y: 2}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 1, Y: 1}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 0, Y: 0}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 1, Y: 0}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 0, Y: 1}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 3, Y: 1}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 3, Y: 2}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 2, Y: 1}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 4, Y: 0}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 4, Y: 1}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 4, Y: 2}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 3, Y: 6}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 0, Y: 3}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 0, Y: 2}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 2, Y: 0}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 3, Y: 0}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 2, Y: 6}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 1, Y: 6}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 1, Y: 7}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 5, Y: 5}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 3, Y: 7}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 5, Y: 4}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 0, Y: 6}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 2, Y: 4}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 0, Y: 4}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 0, Y: 5}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 1, Y: 4}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 0, Y: 7}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 1, Y: 5}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 2, Y: 7}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 6, Y: 4}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 2, Y: 5}},
		core.Move{Side: core.BLACK, Coordinate: core.Coordinate{X: 4, Y: 6}},
		core.Move{Side: core.WHITE, Coordinate: core.Coordinate{X: 4, Y: 7}},
	}
	for _, move := range moves {
		brain.AttemptMove(move, handler)
	}

	brain.PrintGameState()
}
