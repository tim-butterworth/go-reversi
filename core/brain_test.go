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

type TestEventConsumer struct {
	events    []core.Event
	consumers []core.EventConsumer
}

func (consumer *TestEventConsumer) SendEvent(event core.Event) {
	existingEventCount := len(consumer.events)
	nextEvents := make([]core.Event, existingEventCount+1)

	for i, event := range consumer.events {
		nextEvents[i] = event
	}
	nextEvents[existingEventCount] = event

	consumer.events = nextEvents

	for _, c := range consumer.consumers {
		c.SendEvent(event)
	}
}
func (consumer *TestEventConsumer) register(toRegister core.EventConsumer) {
	consumersLength := len(consumer.consumers)
	updatedConsumers := make([]core.EventConsumer, consumersLength+1)

	for i, c := range consumer.consumers {
		updatedConsumers[i] = c
	}
	updatedConsumers[consumersLength] = toRegister

	consumer.consumers = updatedConsumers
}

func NewTestEventConsumer() TestEventConsumer {
	return TestEventConsumer{events: []core.Event{}}
}

type TestCommandRejectHandler struct {
	rejectWasCalled bool
}

func (rejectHandler *TestCommandRejectHandler) InvalidCommand(command core.Command) {
	rejectHandler.rejectWasCalled = true
}
func NewTestCommandRejectHandler() TestCommandRejectHandler {
	return TestCommandRejectHandler{
		rejectWasCalled: false,
	}
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

func Test_InitializeCommand_triggersGameStateUpdate(t *testing.T) {
	testEventConsumer := NewTestEventConsumer()

	brain := core.NewGameBrain(&testEventConsumer)
	testRejectHandler := NewTestCommandRejectHandler()

	brain.Initialize(&testRejectHandler)

	eventCount := len(testEventConsumer.events)
	if !(eventCount == 1) {
		t.Errorf("Expected 1 event, instead got %d", eventCount)
	}

	event := testEventConsumer.events[0]
	if !(event.EventType == core.INITILIZED) {
		t.Error("Should have been an INITIALIZED event")
	}
}

func Test_SecondInitializeCommand_isRejected(t *testing.T) {
	testEventConsumer := NewTestEventConsumer()

	brain := core.NewGameBrain(&testEventConsumer)
	testRejectHandler := NewTestCommandRejectHandler()

	brain.Initialize(&testRejectHandler)
	if testRejectHandler.rejectWasCalled {
		t.Error("First Initialize should not have been rejected")
	}

	brain.Initialize(&testRejectHandler)
	if !testRejectHandler.rejectWasCalled {
		t.Error("Second Initialize should have been rejected")
	}
}

func Test_ValidMove_triggersGameStateUpdate(t *testing.T) {
	testEventConsumer := NewTestEventConsumer()

	brain := core.NewGameBrain(&testEventConsumer)
	testRejectHandler := NewTestCommandRejectHandler()

	brain.Initialize(&testRejectHandler)
	brain.ExecuteCommand(core.NewMoveCommand(core.BLACK, core.Coordinate{X: 2, Y: 4}), &testRejectHandler)

	eventCount := len(testEventConsumer.events)
	if !(eventCount == 2) {
		t.Errorf("Expected 2 events, instead got %d", eventCount)
	}

	event := testEventConsumer.events[1]
	if !(event.EventType == core.MOVED) {
		t.Error("The second event should have been an MOVED event")
	}
}

func Test_ValidMove_doesNotSignalInvalidCommand(t *testing.T) {
	testEventConsumer := NewTestEventConsumer()

	brain := core.NewGameBrain(&testEventConsumer)
	testRejectHandler := NewTestCommandRejectHandler()

	brain.Initialize(&testRejectHandler)
	brain.ExecuteCommand(core.NewMoveCommand(core.BLACK, core.Coordinate{X: 2, Y: 4}), &testRejectHandler)

	if testRejectHandler.rejectWasCalled {
		t.Error("A successful command should not be rejected")
	}
}

func Test_MoveByWrongSide_isRejected(t *testing.T) {
	testEventConsumer := NewTestEventConsumer()

	brain := core.NewGameBrain(&testEventConsumer)
	testRejectHandler := NewTestCommandRejectHandler()

	brain.Initialize(&testRejectHandler)
	brain.ExecuteCommand(core.NewMoveCommand(core.WHITE, core.Coordinate{X: 2, Y: 4}), &testRejectHandler)

	if !testRejectHandler.rejectWasCalled {
		t.Error("Move for wrong side should have been rejected")
	}

	eventCount := len(testEventConsumer.events)
	if !(eventCount == 1) {
		t.Errorf("Expected 1 event, instead got %d", eventCount)
	}

	event := testEventConsumer.events[0]
	if !(event.EventType == core.INITILIZED) {
		t.Errorf("The second event should have been an %s event\n", core.INITILIZED)
	}
}

func Test_MoveToInvalidLocation_isRejected(t *testing.T) {
	testEventConsumer := NewTestEventConsumer()

	brain := core.NewGameBrain(&testEventConsumer)
	testRejectHandler := NewTestCommandRejectHandler()

	brain.Initialize(&testRejectHandler)
	brain.ExecuteCommand(core.NewMoveCommand(core.BLACK, core.Coordinate{X: 0, Y: 0}), &testRejectHandler)

	if !testRejectHandler.rejectWasCalled {
		t.Error("Move to invalid location should be rejected")
	}

	eventCount := len(testEventConsumer.events)
	if !(eventCount == 1) {
		t.Errorf("Expected 1 event, instead got %d", eventCount)
	}

	event := testEventConsumer.events[0]
	if !(event.EventType == core.INITILIZED) {
		t.Errorf("The second event should have been an %s event\n", core.INITILIZED)
	}
}

func Test_ValidMoveBySecondPlayer_isAccepted(t *testing.T) {
	testEventConsumer := NewTestEventConsumer()

	brain := core.NewGameBrain(&testEventConsumer)
	testRejectHandler := NewTestCommandRejectHandler()

	brain.Initialize(&testRejectHandler)
	brain.ExecuteCommand(core.NewMoveCommand(core.BLACK, core.Coordinate{X: 2, Y: 4}), &testRejectHandler)
	brain.ExecuteCommand(core.NewMoveCommand(core.WHITE, core.Coordinate{X: 2, Y: 3}), &testRejectHandler)

	if testRejectHandler.rejectWasCalled {
		t.Error("Should not reject a valid move")
	}

	eventCount := len(testEventConsumer.events)
	if !(eventCount == 3) {
		t.Errorf("Expected 3 events, instead got %d", eventCount)
	}
}

func Test_MoveByWrongSecondPlayer_isRejected(t *testing.T) {
	testEventConsumer := NewTestEventConsumer()

	brain := core.NewGameBrain(&testEventConsumer)
	testRejectHandler := NewTestCommandRejectHandler()

	brain.Initialize(&testRejectHandler)
	brain.ExecuteCommand(core.NewMoveCommand(core.BLACK, core.Coordinate{X: 2, Y: 4}), &testRejectHandler)

	if testRejectHandler.rejectWasCalled {
		t.Error("Should not reject a valid move")
	}

	brain.ExecuteCommand(core.NewMoveCommand(core.BLACK, core.Coordinate{X: 2, Y: 3}), &testRejectHandler)

	if !testRejectHandler.rejectWasCalled {
		t.Error("By a wrong player should be rejected")
	}

	eventCount := len(testEventConsumer.events)
	if !(eventCount == 2) {
		t.Errorf("Expected 2 event, instead got %d", eventCount)
	}
}

func Test_MoveBySecondPlayer_toInvalidLocation_isRejected(t *testing.T) {
	testEventConsumer := NewTestEventConsumer()

	brain := core.NewGameBrain(&testEventConsumer)
	testRejectHandler := NewTestCommandRejectHandler()

	brain.Initialize(&testRejectHandler)
	brain.ExecuteCommand(core.NewMoveCommand(core.BLACK, core.Coordinate{X: 2, Y: 4}), &testRejectHandler)

	if testRejectHandler.rejectWasCalled {
		t.Error("Should not reject a valid move")
	}

	brain.ExecuteCommand(core.NewMoveCommand(core.WHITE, core.Coordinate{X: 0, Y: 0}), &testRejectHandler)

	if !testRejectHandler.rejectWasCalled {
		t.Error("By a wrong player should be rejected")
	}

	eventCount := len(testEventConsumer.events)
	if !(eventCount == 2) {
		t.Errorf("Expected 2 event, instead got %d", eventCount)
	}
}

func Test_Entire_SequenceOfMovesForAGame(t *testing.T) {
	moves := []core.Coordinate{
		{X: 2, Y: 4},
		{X: 2, Y: 5},
		{X: 2, Y: 6},
		{X: 1, Y: 4},
		{X: 0, Y: 4},
		{X: 4, Y: 5},
		{X: 5, Y: 2},
		{X: 4, Y: 2},
		{X: 3, Y: 2},
		{X: 2, Y: 1},
		{X: 3, Y: 1},
		{X: 4, Y: 1},
		{X: 3, Y: 0},
		{X: 1, Y: 5},
		{X: 4, Y: 6},
		{X: 2, Y: 7},
		{X: 3, Y: 6},
		{X: 1, Y: 3},
		{X: 0, Y: 5},
		{X: 3, Y: 5},
		{X: 0, Y: 2},
		{X: 1, Y: 2},
		{X: 2, Y: 3},
		{X: 2, Y: 2},
		{X: 3, Y: 7},
		{X: 0, Y: 3},
		{X: 2, Y: 0},
		{X: 6, Y: 2},
		{X: 7, Y: 2},
		{X: 5, Y: 3},
		{X: 6, Y: 4},
		{X: 5, Y: 4},
		{X: 4, Y: 7},
		{X: 7, Y: 4},
		{X: 7, Y: 5},
		{X: 5, Y: 7},
		{X: 7, Y: 3},
		{X: 6, Y: 3},
		{X: 5, Y: 6},
		{X: 5, Y: 5},
		{X: 6, Y: 5},
		{X: 5, Y: 0},
		{X: 5, Y: 1},
		{X: 4, Y: 0},
		{X: 6, Y: 0},
		{X: 1, Y: 7},
		{X: 1, Y: 1},
		{X: 1, Y: 0},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 6},
		{X: 0, Y: 7},
		{X: 0, Y: 6},
		{X: 7, Y: 6},
		{X: 7, Y: 7},
		{X: 6, Y: 6},
		{X: 6, Y: 7},
		{X: 7, Y: 1},
		{X: 6, Y: 1},
		{X: 7, Y: 0},
	}

	testEventConsumer := NewTestEventConsumer()

	brain := core.NewGameBrain(&testEventConsumer)
	testRejectHandler := NewTestCommandRejectHandler()

	brain.Initialize(&testRejectHandler)

	flip := 1
	for _, move := range moves {
		side := core.BLACK
		if flip == -1 {
			side = core.WHITE
		}
		flip = flip * -1

		brain.ExecuteCommand(core.NewMoveCommand(side, move), &testRejectHandler)
	}

	if testRejectHandler.rejectWasCalled {
		t.Error("All moves should have been valid")
	}
}
