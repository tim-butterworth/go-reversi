package core

type Itterator interface {
	HasNext() bool
	Next() []int
}

type Coordinate struct {
	X int
	Y int
}

type Counter struct {
	values  [][]int
	indexes []int
	current int
	hasMore bool
}

func (counter *Counter) HasNext() bool {
	return counter.hasMore
}

func (counter *Counter) Next() []int {
	max := len(counter.indexes)

	result := make([]int, max)
	for i, v := range counter.indexes {
		result[i] = counter.values[i][v]
	}

	maxIndex := max - 1
	pointer := maxIndex
	for {
		currentIndex := counter.indexes[pointer]
		if currentIndex < len(counter.values[pointer]) - 1 {
			counter.indexes[pointer] = currentIndex + 1
			break
		} else {
			counter.indexes[pointer] = 0
		}

		if pointer > 0 {
			pointer = pointer - 1
		} else {
			counter.hasMore = false
			break
		}
	}

	return result
}

func empty(values [][]int) bool {
	if len(values) == 0 {
		return true
	}

	emptySublist := false
	for _, v := range values {
		if len(v) == 0 {
			emptySublist = true
			break
		}
	}

	return emptySublist
}

func NewCounter(values [][]int) Itterator {
	indexes := make([]int, len(values))
	for i := range indexes {
		indexes[i] = 0
	}

	hasMore := !empty(values)
	return &Counter{
		values:  values,
		indexes: indexes,
		current: 0,
		hasMore: hasMore,
	}
}