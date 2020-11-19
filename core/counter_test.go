package core_test

import (
	"reversi/core"
	"testing"
)

func TestCanInrementOnePlace(t *testing.T) {
	values := []int{0, 1}
	counter := core.NewCounter([][]int{values})

	expectedValues := []int{0, 1}
	for _, expectedValue := range expectedValues {
		actualValueContainer := counter.Next()
		actualvalueLength := len(actualValueContainer)
		if actualvalueLength == 1 {
			actualValue := actualValueContainer[0]
			if expectedValue != actualValue {
				t.Errorf("Expected %d but got %d\n", expectedValue, actualValue)
			}
		} else {
			t.Errorf("Expected a result with length 1 actually got %d\n", actualvalueLength)
		}
		expectedValue = expectedValue + 1
	}

	if counter.HasNext() {
		t.Error("Itterator should be out of values")
	}
}

func TestDoesNotHaveNextForZeroColumns(t *testing.T) {
	counter := core.NewCounter([][]int{})

	if counter.HasNext() {
		t.Error("An empty list should not have a next")
	}
}

func TestDoesNotHaveNextForNil(t *testing.T) {
	counter := core.NewCounter(nil)

	if counter.HasNext() {
		t.Error("An empty list should not have a next")
	}
}

func TestDoesNotHaveNextIfThereIsAnEmptyColumn(t *testing.T) {
	counter := core.NewCounter([][]int{
		[]int{1, 2},
		[]int{},
	})

	if counter.HasNext() {
		t.Error("An empty list should not have a next")
	}
}
